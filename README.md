# gradle-pretty
Formatter to make gradle build output readable, inspired by
[xcpretty](https://github.com/xcpretty/xcpretty).

```bash
# Build and install
go build -o gradle-pretty ./src
install gradle-pretty ~/.local/bin/gradle-pretty

# Use with gradle build
./gradlew build 2>&1 | gradle-pretty
```
