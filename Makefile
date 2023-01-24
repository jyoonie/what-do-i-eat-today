# by putting PHONY list on top, you're telling makefile that you're not asking it to make a file, you're asking it to do something
.PHONY: dbuild dclean drun database

dbuild:
	docker build -t wdiet:latest .

dclean:
	docker stop wdiet && docker rm wdiet && docker image rm wdiet

drun:
	docker run -d -p 8080:8080 --network jynet --name wdiet wdiet

database:
	docker run -d -p 5432:5432 --network jynet --name wdiet_db -e POSTGRES_PASSWORD=secret postgres:alpine