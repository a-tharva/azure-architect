FROM python:3.10-slim-buster

WORKDIR /python-docker

COPY ./repo/backend/requirements.txt requirements.txt
RUN pip3 install -r requirements.txt

COPY ./repo/backend/ .

EXPOSE 8000/tcp

CMD [ "python", "app.py"]