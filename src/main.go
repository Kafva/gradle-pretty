package main

import (
    "bufio"
    "flag"
    "fmt"
    "os"
    "strings"
    "time"
    "golang.org/x/term"
)

type GradleTask struct {
    Name string
    Failed bool
}

type GradleError struct {
    Location string
    Desc string
}

func die(fmtStr string, args... any) {
    fmt.Printf(fmtStr, args ...)
    os.Exit(1)
}

func taskLog(task GradleTask, width int) {
    var result string
    if task.Failed {
        result = " \033[91m\033[0m"
    } else {
        result = ""
    }
    msg := fmt.Sprintf("\r\033[93m▸\033[0m %s%s", task.Name, result)
    spaces := strings.Repeat(" ", width - len(msg))
    print(msg + spaces) // No newline
}

func parseBuildLog(noLogfile bool, logfile string) (
    tasks []GradleTask,
    errors []GradleError,
    timeTaken int64) {
    var f *os.File
    var err error
    if !noLogfile {
        f, err = os.OpenFile(logfile,
                             os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
        if err != nil {
            die("Error opening %s: %s\n", logfile, err)
        }
        defer f.Close()
    }
    cwd, err := os.Getwd()
    if err != nil {
        die("Error reading current working directory: %s\n", err)
    }
    cwd = "file://" + cwd + "/"
    termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        die("Error reading terminal size: %s\n", err)
    }

    scanner := bufio.NewScanner(os.Stdin)

    startTime := time.Now().Unix()
    for scanner.Scan() {
        rawline := scanner.Text()
        line := strings.TrimSpace(rawline)

        if strings.HasPrefix(line, "> Task") {
            spl := strings.Split(line, " ")
            if len(spl) < 4 {
                continue
            }
            task := GradleTask { spl[2], spl[3] == "FAILED" }
            tasks = append(tasks, task)

            taskLog(task, termWidth)

        } else if strings.HasPrefix(line, "e:") {
            // Source code errors have an 'e:' prefix
            spl := strings.Split(line, " ")
            if len(spl) < 2 {
                continue
            }
            location, _ := strings.CutPrefix(spl[1], cwd)
            desc := strings.Join(spl[2:], " ")
            err := GradleError { location, desc }
            errors = append(errors, err)

        }

        if !noLogfile {
            f.WriteString(rawline + "\n")
        }
    }
    endTime := time.Now().Unix()
    timeTaken = endTime - startTime
    println()

    return tasks, errors, timeTaken
}

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
        fmt.Fprintf(os.Stderr, "EXAMPLE:\n")
        fmt.Fprintf(os.Stderr, "  ./gradlew build 2>&1 | %s\n", os.Args[0])
    }
    noLogfile := flag.Bool("N", false, "Do not save a copy of the complete build log")
    logfile := flag.String("l", "build.log", "Path to save complete build log in")
    flag.Parse()

    tasks, errors, timeTaken := parseBuildLog(*noLogfile, *logfile)

    for _,task := range tasks {
        if !task.Failed {
            continue
        }

        // Dump errors if at least one task failed
        for _,err := range errors {
            fmt.Printf("%s: %s\n", err.Location, err.Desc)
        }

        fmt.Printf("\033[91mBUILD FAILED\033[0m in %ds\n", timeTaken)
        if ! *noLogfile {
            fmt.Printf("See %s for more information\n", *logfile)
        }
        os.Exit(1)
    }

    fmt.Printf("\033[92mBUILD SUCCESSFUL\033[0m in %ds\n", timeTaken)
}
