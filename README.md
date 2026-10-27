# Asgardeo CLI releases

This branch exists only to host releases of the Asgardeo CLI (`asg`), tagged
`asg-v*` in this repository, and the scripts that install them. It contains no
source code.

The `asg` CLI is a standalone tool for managing Asgardeo organizations. It is
not part of the WSO2 CLI (`ws`) and is not installed by it.

## Install

macOS and Linux:

```sh
curl -fsSL https://raw.githubusercontent.com/wso2/wso2-cli/release/asgardeo-cli/scripts/install.sh | bash
```

Windows:

```powershell
iwr https://raw.githubusercontent.com/wso2/wso2-cli/release/asgardeo-cli/scripts/install.ps1 -useb | iex
```

By default both install the newest stable `asg-v*` release. Set
`ASG_PRERELEASE=true` to include release candidates. Either way the binary is
verified against the `checksums.txt` published beside it, and no administrator
rights are needed. Running either again updates an existing install. Each
script lists the variables it reads at the top, including how to pin a version.
