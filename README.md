# go-todo-api
A to do list application that allows for adding tasks, completing tasks, moving them to different tables, viewing both tables and more. This originally was supposed to be an API but that changed and it was turned in a "full-stack" app using GO templates.

---
### GitHub Actions
The workflows in the project are build and push changes to Docker hub while updating another repositories tags with the newest image which is pulled down to my k3s cluster via ArgoCD, completing the entire CI/CD pipeline. More about the building of the pipeline can be found on [my blog](https://khenry.substack.com/p/the-hyperbolic-chamber-12182023).

![CI/CD architecture](https://github.com/kwehen/go-todo-api/assets/110314567/3b36f848-8874-49c6-94f6-fce4b2f99236)

---
Update:
I've added OAUTH and changed the database schema. I'm hoping to make this publically available for some friends and whoever else to use. I'm tightening up security before releasing, and I hope to add features that users see are needed.
