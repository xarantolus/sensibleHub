rm sensibleHub.zip || true 
rm sensibleHub || true 
rm sensibleHub.exe || true 

# Use first argument for name if possible, fallback to sensibleHub.zip
NAME=${1:-sensibleHub.zip}    

go build -v -mod vendor -ldflags "-s -w"

zip -r "$NAME" LICENSE README.md templates/ assets/ config.json sensibleHub*

echo "You can now extract $NAME somewhere on the target system and run the executable"