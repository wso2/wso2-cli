Prerequiste

Install Choreo CLI via ```curl -o- https://cli.choreo.dev/install.sh | bash```
Login to Choreo with ```choreo login```


1. Create a fork of [https://github.com/wso2/choreo-samples]() (Skip to #3 if you already have a clone locally)
2. Clone the fork to you local machine.
   ```
        gh repo clone kaviththiranga/choreo-samples
   ```
3. Create a mono repo project from the repository

    ```
        choreo create project --repo=. my-reading-list-app 
    ```
4. Create a Component for the BE aplication
    ```
        cd reading-list-app/reading-list-service/
        choreo create component --sub-path=. --project=my-reading-list-app my-reading-list-svc
    ```
5. build the BE Component
    ```
        choreo build --project=my-reading-list-app --component=my-reading-list-svc
    ```

6. Deploy into dev (Dev is the deafault env, main is the default deployment track)

    ```
        choreo deploy --project=my-reading-list-app --component=my-reading-list-svc
    ```

7. Try it out by going into Choreo Console
   (alternatively, choreo describe component to get endpoints, choreo get-token to get a test token and curl)

8. Promote to Production

    ```
         choreo deploy --env=production --project=my-reading-list-app --component=my-reading-list-svc
    ```

9. Configure CORS for production endpoint via Choreo Console API Management Page (need to discuss an alternative if possible), Enable pass security context option to pass token to the BE

10. Create the Component for the FE Application

    ```
        cd ../reading-list-front-end-with-managed-auth/
        choreo create component --sub-path=. --project=my-reading-list-app my-reading-list-web-app
        choreo build --project=my-reading-list-app --component=my-reading-list-web-app
    ```

11. Create Connection To BE
    ```
        choreo create dependency --type=connection --project=my-reading-list-app --component=my-reading-list-web-app --service=my-reading-list-svc my-webapp-to-backend-con
    ```

12 Deploy
    ```
        choreo deploy --project=my-reading-list-app --component=my-reading-list-web-app 

    ```

13. Test the webApp

14. Promote

    ```
        choreo deploy --env=production --project=my-reading-list-app --component=my-reading-list-web-app 
    ```
