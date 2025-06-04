cd test
python main.py
cd ../

cd cmd/cli
go build -o ../../bin/timetabler .
cd ../../

cp config.json bin
