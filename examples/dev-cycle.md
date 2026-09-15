```bash
$ choreo component deploy -n MyComponent -o MyTechOrg -p ProjectB
```

```
🚀 Deploying Component 'MyComponent'

⚙️ Fetching configurations for the component...

✅ Configuration retrieved successfully.
🔍 Initiating deployment...

🔄 Deployment in progress...
⏳ This might take a moment.

📈 To check logs for this component, use:
choreo component logs -n MyComponent -o MyTechOrg -p ProjectB

🔗 Your component's accessible endpoint:
https://mycomponent.mytechorg.projectb.choreo.dev

```

```
$ git add -u
$ git commit -m "Change Service Logic"
$ git push origin
```

```
$ choreo component deploy -n MyComponent -o MyTechOrg -p ProjectB
```

```

🔄 Rebuilding Component 'MyComponent'

📦 Building the updated codebase...

✅ Build completed successfully.

🚀 Initiating redeployment...

🔄 Redeployment in progress...
⏳ This might take a moment.

📈 To check logs for this component, use:
choreo component logs -n MyComponent -o MyTechOrg -p ProjectB

🔗 Your component's accessible endpoint:
https://mycomponent.mytechorg.projectb.choreo.dev

```

```
$ choreo component logs -n MyComponent -o MyTechOrg -p ProjectB
📊 Component Logs for 'MyComponent'

🔍 Fetching logs...

[timestamp] INFO: Component initialized.
[timestamp] DEBUG: Handling incoming request...

[timestamp] ERROR: An error occurred in component logic.
[timestamp] INFO: Component shutdown.

Logs retrieved successfully.
```
