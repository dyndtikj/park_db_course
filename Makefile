ALL_PKG = ./...

build:
	docker image rm --force park-forum
	docker build --no-cache -t park-forum .

run:
	docker run --memory 2G --log-opt max-size=5M --log-opt max-file=3 -p 5000:5000 -p 5432:5432  -t park-forum

run-all-docker: build run

run-func-test:
	curl -vvv -X POST http://localhost:5000/api/service/clear
	./bin/tp-dbms-forum-test-tool func -u http://localhost:5000/api/ -r report.html

run-perf-fill:
	curl -vvv -X POST http://localhost:5000/api/service/clear
	./bin/tp-dbms-forum-test-tool fill -u http://localhost:5000/api/ --timeout=900

run-perf-test:
	./bin/tp-dbms-forum-test-tool perf -u http://localhost:5000/api/  --duration=600 --step=60

run-all-tests:
	# очищаем бд
	curl -vvv -X POST http://localhost:5000/api/service/clear
	make run-func-test
	# заполнение тестовыми данными
	make run-perf-fill
	# тестируем производительность
	make run-perf-test


init-test-tool:
	go get -u -v github.com/mailcourses/technopark-dbms-forum@master
	go build github.com/mailcourses/technopark-dbms-forum
	mv technopark-dbms-forum ./bin/tp-dbms-forum-test-tool

generate:
	find ./ -name "*_easyjson.go" -exec rm -rf {} \;
	go generate ${ALL_PKG}