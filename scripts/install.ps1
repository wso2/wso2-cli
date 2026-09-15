$BINARY_NAME = "wso2-integration-platform"
$REPO = "wso2/integration-platform-tools"

function Get-Architecture {
    switch ([Environment]::Is64BitOperatingSystem) {
        $true { return "amd64" }
        $false { return "386" }
    }
}

function Get-Version {
    param($version)
    if ($version) {
        return $version
    }

    if ($env:LAST_RELEASE -eq "true") {
        $response = Invoke-WebRequest -Method Get -Uri "https://api.github.com/repos/$REPO/releases"
        $object = $response.Content | ConvertFrom-Json
        foreach ($release in $object) {
            if ($release.prerelease -eq $true) {
                return $release.tag_name
            }
        }
    }

    $url = "https://github.com/$REPO/releases/latest"
    $redirect_response = Invoke-WebRequest -Method Get -Uri $url -MaximumRedirection 0 -ErrorAction SilentlyContinue
    if ($redirect_response.StatusCode -eq 302) {
        return $redirect_response.Headers.Location.Split("/")[-1]
    }
}

function Main {
    param($version)
    $ARCH = Get-Architecture
    $VERSION = Get-Version $version
    $INSTALL_DIR = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$HOME\.wso2-integration-platform\bin" }
    $EXE_PATH = "$INSTALL_DIR\$BINARY_NAME.exe"
    $ZIP_NAME = "$BINARY_NAME-$VERSION-windows-$ARCH.zip"
    $ZIP_PATH = "$env:TEMP\$ZIP_NAME"

    if (!(Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
    }

    $prevProgressPreference = $ProgressPreference
    try {
        if ($PSVersionTable.PSVersion.Major -lt 7) {
            $ProgressPreference = "SilentlyContinue"
        }

        $downloadUri = "https://github.com/$REPO/releases/download/$VERSION/$ZIP_NAME"
        Write-Output "Downloading $BINARY_NAME $VERSION..."
        Write-Output "URL: $downloadUri"
        Invoke-WebRequest $downloadUri -OutFile $ZIP_PATH
    } finally {
        $ProgressPreference = $prevProgressPreference
    }

    Write-Output "Extracting..."
    if (Get-Command Expand-Archive -ErrorAction SilentlyContinue) {
        Expand-Archive $ZIP_PATH -Destination $INSTALL_DIR -Force
    } else {
        Write-Error "Expand-Archive not available. Please extract $ZIP_PATH manually."
        exit 1
    }
    Remove-Item $ZIP_PATH

    $User = [EnvironmentVariableTarget]::User
    $Path = [Environment]::GetEnvironmentVariable("Path", $User)
    if (!(";$Path;".ToLower() -like "*;$INSTALL_DIR;*".ToLower())) {
        [Environment]::SetEnvironmentVariable('Path', "$Path;$INSTALL_DIR", $User)
        $Env:Path += ";$INSTALL_DIR"
    }

    Write-Output ""
    Write-Output "$BINARY_NAME $VERSION installed to $EXE_PATH"
    Write-Output ""
    Write-Output "Get started:"
    Write-Output "  $BINARY_NAME login"
    Write-Output ""
    Write-Output "Use as MCP server with Claude Code:"
    Write-Output "  claude mcp add wso2-integration-platform -- $BINARY_NAME start-mcp-server"
}

Main $args[0]
