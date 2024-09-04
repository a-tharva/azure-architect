Instructions
- docker build -t flask_image:v9 .
- docker container run -dp 80:8000 -t flask_image:v9