## 📖 Swagger Documentation

To view Swagger UI locally:
````bash
go install github.com/swaggo/swag/cmd/swag@latest
````

---

**С шагами:**
````markdown
## 📖 API Documentation

> Swagger UI is not hosted. To view docs locally, follow these steps:

1. Install the `swag` CLI:
   
   go install github.com/swaggo/swag/cmd/swag@latest
   
2. Generate docs:

   swag init -g cmd/app/main.go

3. Run the server and open http://localhost:8080/swagger/index.html
````