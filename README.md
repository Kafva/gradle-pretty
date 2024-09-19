# gradle-pretty
Formatter to make gradle build output readable, inspired by
[xcpretty](https://github.com/xcpretty/xcpretty). The default build log is too
noisy, it is a pain to find the actual errors in the output. With
`gradle-pretty` we only show the current task and source code errors on
failure, no more, no less.

```bash
go install github.com/Kafva/gradle-pretty
./gradlew build 2>&1 | gradle-pretty
```
