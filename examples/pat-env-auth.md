# PAT Environment Variable Authentication

This document describes how to use Personal Access Token (PAT) authentication via environment variable with the WSO2 Integration Platform CLI.

## Overview

The CLI supports automatic authentication using a Personal Access Token (PAT) stored in the `WSO2IP_PAT` environment variable. This allows users to authenticate without having to run `wso2-integration-platform login --with-token` manually.

## Usage

### Setting the Environment Variable

Set the `WSO2IP_PAT` environment variable with your Personal Access Token:

```bash
export WSO2IP_PAT="your-personal-access-token-here"
```

### How It Works

1. When the CLI needs to check if a user is authenticated, it first checks for stored authentication (browser login or previous PAT login)
2. If no stored authentication is found, it checks for the `WSO2IP_PAT` environment variable
3. If the environment variable is set, it attempts to authenticate using that PAT
4. If authentication is successful, the user info and token are stored for future use
5. If authentication fails, the CLI falls back to prompting the user to login

### Benefits

- **Automation-friendly**: Perfect for CI/CD pipelines and automated scripts
- **No interactive prompts**: Eliminates the need for manual token input
- **Secure**: Tokens are stored securely in environment variables
- **Fallback support**: Still works with existing browser-based and manual PAT authentication

### Example

```bash
# Set your PAT
export WSO2IP_PAT="your-token-here"

# Now you can run any CLI command without manual authentication
wso2-integration-platform list projects
wso2-integration-platform list components
wso2-integration-platform create project my-project
```

### Security Considerations

- Store the `WSO2IP_PAT` environment variable securely
- Use environment-specific tokens when possible
- Rotate tokens regularly
- Never commit tokens to version control

### Troubleshooting

If authentication fails with the environment variable:

1. Verify the token is valid and not expired
2. Check that the token has the necessary permissions
3. Ensure the environment variable is set correctly: `echo $WSO2IP_PAT`
4. Try running `wso2-integration-platform login --with-token` manually to test the token

### Compatibility

This feature is compatible with:
- All existing CLI commands
- Both browser-based and manual PAT authentication
- CI/CD environments
- All supported operating systems
