# Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
#
# WSO2 LLC. licenses this file to you under the Apache License,
# Version 2.0 (the "License"); you may not use this file except
# in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# Installs the Asgardeo CLI (asg) on Windows.
#
#   iwr <install url> -useb | iex                              # newest stable release
#   $env:ASG_VERSION = 'asg-v0.1.0'; iwr <install url> -useb | iex   # a pinned release
#
# The Windows counterpart of install.sh, and it behaves the same way: the newest
# asg-v* release is found by listing releases and comparing versions, never by
# asking GitHub for Latest, which belongs to the WSO2 CLI. The binary is verified
# against checksums.txt before it is installed, and running the script again is
# safe: an install that already reports the version is left alone, an older one
# is replaced.
#
# Variables it reads:
#
#   ASG_VERSION            A release tag to install.
#   ASG_PRERELEASE=true    Resolve the newest release including candidates.
#   ASG_INSTALL_DIR        Where to put asg.exe. Default
#                          %LOCALAPPDATA%\Programs\asg. When set, the user PATH
#                          is not edited.
#   ASG_NO_PROFILE=1       Install without editing the user PATH.
#   ASG_RELEASE_BASE_URL   Where releases are downloaded from, and where they
#   ASG_RELEASE_API_URL    are listed. Overridden by tests.

function Get-AsgVersionKey([string] $Version) {
    $core, $pre = $Version -split '-', 2
    $n = $core -split '\.'
    $key = '{0:D10}.{1:D10}.{2:D10}.' -f [long]$n[0], [long]$n[1], [long]$n[2]
    if (-not $pre) { return $key + '1' }
    $parts = foreach ($part in $pre -split '\.') {
        if ($part -match '^\d+$') { '0' + ('{0:D10}' -f [long]$part) } else { '1' + $part }
    }
    return $key + '0' + ($parts -join ',')
}

# Add-AsgToUserPath appends a directory to the user PATH, unless it is already
# there, and prints what it did.
#
# The value is read and written in the registry directly. The user PATH is
# normally stored as an expandable string, so an entry like %USERPROFILE%\bin
# stays a reference. [Environment]::GetEnvironmentVariable hands it back with
# every variable already expanded, and SetEnvironmentVariable writes a plain
# string, so going through them would rewrite every such entry as a fixed path
# and change the value's type for good. Here it is read unexpanded and written
# back with the type it already had.
function Add-AsgToUserPath([string] $Directory) {
    $key = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
    try {
        $raw = [string]$key.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
        $kind = if ($raw) { $key.GetValueKind('Path') } else { [Microsoft.Win32.RegistryValueKind]::ExpandString }

        $wanted = $Directory.TrimEnd('\')
        foreach ($entry in $raw -split ';') {
            $expanded = [Environment]::ExpandEnvironmentVariables($entry.Trim()).TrimEnd('\')
            if ($expanded -and $expanded -ieq $wanted) {
                return 'already on PATH'
            }
        }

        $updated = if ($raw.TrimEnd(';')) { "$($raw.TrimEnd(';'));$Directory" } else { $Directory }
        $key.SetValue('Path', $updated, $kind)
    }
    finally {
        $key.Close()
    }

    # A registry write does not tell running programs the environment changed,
    # so a terminal opened from Explorer afterwards would not see it. Setting
    # any user variable through .NET broadcasts that change; clearing one that
    # was never set changes nothing else.
    [Environment]::SetEnvironmentVariable('ASG_INSTALLER_NOTIFY', $null, 'User')

    $env:Path = "$env:Path;$Directory"
    return "added $Directory to your user PATH; open a new terminal"
}

function Install-Asg {
    $ErrorActionPreference = 'Stop'
    # Invoke-WebRequest renders a progress bar that slows Windows PowerShell 5.1
    # downloads by an order of magnitude.
    $ProgressPreference = 'SilentlyContinue'
    [Net.ServicePointManager]::SecurityProtocol =
        [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

    $repo = 'wso2/wso2-cli'
    $prefix = 'asg-v'
    $baseUrl = if ($env:ASG_RELEASE_BASE_URL) { $env:ASG_RELEASE_BASE_URL } else { "https://github.com/$repo/releases/download" }
    $apiUrl = if ($env:ASG_RELEASE_API_URL) { $env:ASG_RELEASE_API_URL } else { "https://api.github.com/repos/$repo/releases?per_page=100" }

    # A 32-bit PowerShell on 64-bit Windows reports x86 here and the real
    # architecture in PROCESSOR_ARCHITEW6432.
    $cpu = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
    $arch = switch ($cpu) {
        'AMD64' { 'amd64' }
        'ARM64' { 'arm64' }
        default { throw "No asg release for Windows/$cpu. Download a binary from https://github.com/$repo/releases" }
    }
    $asset = "asg-windows-$arch.exe"

    if ($env:ASG_VERSION) {
        if ($env:ASG_VERSION -notmatch "^$([regex]::Escape($prefix))\d") {
            throw "A release tag looks like ${prefix}0.1.0 (got $($env:ASG_VERSION))."
        }
        $tag = $env:ASG_VERSION
    }
    else {
        try {
            $releases = Invoke-RestMethod -Uri $apiUrl -Headers @{ Accept = 'application/vnd.github+json' }
        }
        catch {
            throw "Could not list releases from GitHub (network, or the unauthenticated API rate limit). Set ASG_VERSION to skip this step."
        }
        $wantPre = $env:ASG_PRERELEASE -eq 'true'
        # What decides a candidate is the Pre-release flag set when publishing,
        # not a suffix in the tag name.
        $candidates = @($releases | Where-Object {
                -not $_.draft -and ($wantPre -or -not $_.prerelease) -and
                $_.tag_name -match '^asg-v\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$'
            })
        # Ranked by the precedence rules of semantic versioning, with the same
        # key install.sh builds: numbers zero-padded, 1 for a release or 0 for a
        # candidate, then each suffix part as 0 and a padded number or 1 and the
        # word, joined by a comma. The keys are compared by byte order on
        # purpose. Sort-Object compares strings by culture, which ranks symbols
        # before letters and would undo the encoding.
        $byKey = @{}
        foreach ($release in $candidates) {
            $byKey[(Get-AsgVersionKey $release.tag_name.Substring($prefix.Length))] = $release
        }
        $keys = [string[]]@($byKey.Keys)
        [Array]::Sort($keys, [StringComparer]::Ordinal)
        $newest = if ($keys.Count) { $byKey[$keys[-1]] } else { $null }
        if (-not $newest) {
            throw "No published ${prefix}* release found in $repo. Set ASG_VERSION, or ASG_PRERELEASE=true for a release candidate."
        }
        $tag = $newest.tag_name
    }

    $dir = if ($env:ASG_INSTALL_DIR) { $env:ASG_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\asg' }
    $target = Join-Path $dir 'asg.exe'

    if (Test-Path $target) {
        $have = ''
        try {
            $line = & $target --version 2>$null | Select-Object -First 1
            if ($line -match '^asg version (\S+)') { $have = $Matches[1] }
        }
        catch { }
        if ($have -and $have -eq $tag.Substring($prefix.Length)) {
            Write-Host "$tag is already installed at $target"
            return
        }
    }

    Write-Host "Installing $tag for windows-$arch."
    $work = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
    New-Item -ItemType Directory -Path $work | Out-Null
    try {
        $binary = Join-Path $work $asset
        $sums = Join-Path $work 'checksums.txt'
        Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$tag/$asset" -OutFile $binary
        Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$tag/checksums.txt" -OutFile $sums

        $entry = Get-Content $sums | Where-Object {
            $name = ($_ -split '\s+')[1]
            $name -eq $asset -or $name -eq "*$asset"
        } | Select-Object -First 1
        if (-not $entry) {
            throw "checksums.txt for $tag does not list $asset, so the download cannot be verified."
        }
        $expected = ($entry -split '\s+')[0].ToLowerInvariant()
        $actual = (Get-FileHash -Algorithm SHA256 -Path $binary).Hash.ToLowerInvariant()
        if ($expected -ne $actual) {
            throw "Checksum mismatch for $asset`n  expected $expected`n  actual   $actual`nRefusing to install a binary that failed verification."
        }

        New-Item -ItemType Directory -Force -Path $dir | Out-Null
        Move-Item -Force -Path $binary -Destination $target
    }
    finally {
        Remove-Item -Recurse -Force -Path $work -ErrorAction SilentlyContinue
    }

    if ($env:ASG_INSTALL_DIR -or $env:ASG_NO_PROFILE -eq '1') {
        $pathNote = "not on PATH; add $dir to it"
    }
    else {
        $pathNote = Add-AsgToUserPath $dir
    }

    Write-Host "$tag installed at $target (checksum verified; $pathNote)"
}

# Everything runs from here, so a download cut off half way through
# `iwr | iex` defines a function and runs nothing.
Install-Asg
