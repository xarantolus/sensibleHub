rm sensibleHub.zip || true 
rm sensibleHub || true 
rm sensibleHub.exe || true 

go build -v -mod vendor -ldflags "-s -w"

zip -r sensibleHub.zip templates/ assets/ config.json sensibleHub*

echo "You can now extract sensibleHub.zip somewhere on the target system and run the executable"