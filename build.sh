cd test
python main.py
cd ../

cp config.json bin

cd cmd/cli
go build -o ../../bin/timetabler .
cd ../../
