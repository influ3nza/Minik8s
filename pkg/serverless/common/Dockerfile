FROM python:latest

COPY . .

# set workdir
WORKDIR /app

EXPOSE 8080

# copy file into workdir
copy . /app
# install dependencies from requirements.txt
RUN pip3 install --default-timeout=60 --no-cache-dir -r ./requirements.txt

# set env
ENV PATH="/usr/local/bin:${PATH}"

# run server.py, start the function server
CMD ["python3", "server.py"]