```
$ cd ~/git/abc-aquarium-store/store-api

$ ls
Ballerina.toml		Dependencies.toml	 	service.bal		target	tests

$ choreo component create
📦 Creating a New Choreo Component

It seems you're about to create a new Choreo component from the source code in this directory.

🌐 Retrieving your organizations...

🔍 You are part of the following organizations:
1. MyTechOrg
2. DevOpsCo
3. InnovativeApps

❓ In which organization would you like to create this component? (Enter the number):
1

🔍 Retrieving projects in organization 'MyTechOrg'...

📚 Available projects:
1. ProjectA
2. ProjectB
3. SideProject
4. create a new project

❓ In which project would you like to create this component? (Enter the number):
2

💡 Let's configure your new component:
🔨 Component Name (default: MyComponent): [Press Enter for default]

🎯 Component Type (default: Service)
Choose from: Service, WebApp, Manual Trigger, Webhook, etc
More about component types: [Link to documentation]
Enter the component type: [Press Enter for default]

❓ Do you want to deploy this component right away? (yes/no): yes

🔧 Creating the new component in project 'ProjectB' of organization 'MyTechOrg'...

✅ Component 'MyComponent' created successfully!

🚀 Starting deployment of the component...

🔍 Deployment in progress...
⏳ This might take a moment.

📈 To check logs for this component, use:
choreo component logs -n MyComponent -o MyTechOrg -p ProjectB

📊 To get the status of your component, use:
choreo component status -n MyComponent -o MyTechOrg -p ProjectB

🔗 Your service component's accessible endpoint:
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
