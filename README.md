# My 1st Go Project
Originally an API, now a full stack app written in GO, using Gin and GO templating. 
---
This is a to-do list application hosted on a k3s cluster in my homelab and exposed for the public internet to use. Let me know about features, vulnerabilities, and anything else you think of that can be added to the application. I am very appreciative of the feedback.

---
### GitHub Actions
The CI workflow in the project builds and pushes changes to Docker hub while updating another repositories tags with the newest image which is pulled down to my k3s cluster via ArgoCD, completing the entire CI/CD pipeline. More about the building of the pipeline can be found on [my blog](https://khenry.substack.com/p/the-hyperbolic-chamber-12182023).

![CI/CD architecture](https://github.com/kwehen/go-todo-api/assets/110314567/3b36f848-8874-49c6-94f6-fce4b2f99236)
The TrueNAS CI builds and pushes changes to a different docker repo which hosts the production image of the application.

### Why TrueNAS? 
Originally the app was hosted and run on the k3s cluster. Because of stability issues when syncing application through ArgoCD (not an HA cluster), it is much more reliable for the application to run on TrueNAS (which runs k3s under the hood) and configure ingress, TLS, etc. from via the UI. 

---
## What's Next?
Give me some feedback and I will implement it if feasible! 
