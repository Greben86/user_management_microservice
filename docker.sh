# сборка контейнера
docker build -t go_service_backend .
# запуск контейнера
docker run -d -p 8081:8081 go_service_backend