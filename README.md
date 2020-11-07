# Department of Labor Form 503 (Voluntary Self‐Identification of Disability Form)

An app for HR dept. to collect data from the Voluntary Self‐Identification of Disability Form(503)
More info on https://www.dol.gov/agencies/ofccp/self-id-forms

Employees data is in a JSON file. 
Employees answers is stored into an sqlite database.
The app connects to an AD server for authentication.
The answers can be retrived all at once (`/data`) or daily (`/data?day=2020-11-30`).

Not part of this repo are two files:

Dockerfile
```Dockerfile
FROM centos:8
WORKDIR /home
RUN dnf upgrade -y && dnf clean all
COPY ./server ./
COPY ./templates ./templates
CMD /home/server
```

docker-compose.yml
```yml
version: '3'

services:
  hrform-503:
    build: .
    container_name: form503
    volumes:
      - ./data:/home/data:Z
    environment:
      - TZ=America/New_York
      - FORM503HOST=0.0.0.0
      - FORM503PORT=8080
      - FORM503DB=data/db.sqlite
      - FORM503EMPLDB=data/employees.json
      - FORM503ADSERVER=....
      - FORM503ADBASEDN=....
    ports:
      - "8080:8080"
```

