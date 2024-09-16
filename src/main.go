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
    Result string
}

type GradleError struct {
    Location string
    Desc string
}

func taskLog(task GradleTask, width int) {
    var result string
    if task.Result == "FAILED" {
        result = " \033[31mFAILED\033[0m"
    } else {
        result = ""
    }
    msg := fmt.Sprintf("\r\033[33m▸\033[0m %s%s", task.Name, result)
    spaces := strings.Repeat(" ", width - len(msg))
    print(msg + spaces) // No newline
}

func main() {
    flag.Usage = func() {
            fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
            flag.PrintDefaults()
            fmt.Fprintf(os.Stderr, "Example:\n")
            fmt.Fprintf(os.Stderr, "  ./gradlew build 2>&1 | %s\n", os.Args[0])
    }
    verbose := flag.Bool("v", false, "Verbose output")
    flag.Parse()

    if *verbose {
        println("Starting...")
    }

    cwd, err := os.Getwd()
    if err != nil {
        fmt.Printf("Error reading current working directory: %s", err)
        os.Exit(1)
    }
    cwd = "file://" + cwd + "/"
    termWidth, _, err := term.GetSize(1)
    if err != nil {
        fmt.Printf("Error reading terminal size: %s", err)
        os.Exit(1)
    }

    var tasks []GradleTask
    var errors []GradleError

    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        if strings.HasPrefix(line, "> Task") {
            spl := strings.Split(line, " ")
            if len(spl) < 4 {
                continue
            }
            task := GradleTask { spl[2], spl[3] }
            tasks = append(tasks, task)

            taskLog(task, termWidth)
            time.Sleep(200 * time.Millisecond)

        } else if strings.HasPrefix(line, "e:") {
            spl := strings.Split(line, " ")
            if len(spl) < 2 {
                continue
            }
            location, _ := strings.CutPrefix(spl[1], cwd)
            desc := strings.Join(spl[2:], " ")
            err := GradleError { location, desc }
            errors = append(errors, err)
        }
    }
    println()

    // Dump errors if at least one task failed
    for _,task := range tasks {
        if task.Result == "FAILED" {
            for _,err := range errors {
                fmt.Printf("%s: %s\n", err.Location, err.Desc)
            }
            break
        }
    }
}
