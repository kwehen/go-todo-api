# go-todo-api
A to do list application that allows for adding tasks, completing tasks, moving them to different tables, viewing both tables and more. This originally was supposed to be an API but that changed and it was turned in a "full-stack" app using GO templates.

---
### GitHub Actions
The workflows in the project are build and push changes to Docker hub while updating another repositories tags with the newest image which is pulled down to my k3s cluster via ArgoCD, completing the entire CI/CD pipeline. More about the building of the pipeline can be found on [my blog](https://khenry.substack.com/p/the-hyperbolic-chamber-12182023).
![CI/CD architecture](https://github.com/kwehen/go-todo-api/assets/110314567/3b36f848-8874-49c6-94f6-fce4b2f99236)

---
The rest of the code is what went into making this project work. SQL files for creating the tables, Dockerfile for testing the creation of the image, and more. The application is hosted internally. But if you want to see it, reach out and I will let you take a look. 
P.S. If I ever feel like adding auth and database sessions I will make this a pubic service. PR coming...
