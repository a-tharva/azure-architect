Instructions
- docker build -t flask_image:v1 .
- docker container run -dp 80:8000 -t flask_image:v1