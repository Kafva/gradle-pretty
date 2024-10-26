package main

import (
    "bufio"
    "flag"
    "fmt"
    "os"
    "strings"
    "time"
)

type GradleTask struct {
    Name string
    Failed bool
}


type GradleIssue struct {
    Location string
    Desc string
    IsError bool
}

type Config struct {
    NoLogfile *bool
    Logfile *string
    NoWarnings *bool
}

func die(fmtStr string, args... any) {
    fmt.Printf(fmtStr, args ...)
    os.Exit(1)
}

func taskLog(task GradleTask) {
    var result string
    if task.Failed {
        result = " \033[91m\033[0m"
    } else {
        result = ""
    }
    msg := fmt.Sprintf("\033[93m▸\033[0m %s%s", task.Name, result)
    print("\r\033[2K")  // Clear line
    println(msg)        // Print with newline
}

func buildOk(tasks []GradleTask, issues []GradleIssue) bool {
    if len(tasks) == 0 {
        return false
    }

    for _,issue := range issues {
        if issue.IsError {
            return false
        }
    }

    for _,task := range tasks {
        if task.Failed {
            return false
        }
    }

    return true
}

func parseBuildLog(cfg *Config) (
    tasks []GradleTask,
    issues []GradleIssue,
    timeTaken int64) {
    var f *os.File
    var cwd string
    var err error
    if !*cfg.NoLogfile {
        f, err = os.OpenFile(*cfg.Logfile,
                             os.O_TRUNC | os.O_CREATE | os.O_WRONLY, 0644)
        if err != nil {
            die("Error opening %s: %s\n", *cfg.Logfile, err)
        }
        defer f.Close()
    }

    cwd, err = os.Getwd()
    if err != nil {
        die("Error reading current working directory: %s\n", err)
    }
    if target, err := os.Readlink(cwd); err == nil {
        cwd = "file://" + target + "/"
    } else {
        cwd = "file://" + cwd + "/"
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
            name, _ := strings.CutPrefix(spl[2], ":")
            task := GradleTask { name, spl[3] == "FAILED" }
            tasks = append(tasks, task)
            if len(tasks) > 1 {
                // After the first log line, always move back up one line
                // before printing anew
                print("\033[A")
            }
            taskLog(task)

        } else {
            isError := strings.HasPrefix(line, "e:")
            if isError || (!*cfg.NoWarnings && strings.HasPrefix(line, "w:")) {
                // Source code errors and warnings
                spl := strings.Split(line, " ")
                if len(spl) < 2 {
                    continue
                }
                location, _ := strings.CutPrefix(strings.TrimSpace(spl[1]), cwd)
                desc := strings.Join(spl[2:], " ")
                issue := GradleIssue { location, desc, isError }
                issues = append(issues, issue)
            }
        }

        if !*cfg.NoLogfile {
            f.WriteString(rawline + "\n")
        }
    }
    endTime := time.Now().Unix()
    timeTaken = endTime - startTime

    return tasks, issues, timeTaken
}

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "Prettify the gradle build log:\n")
        fmt.Fprintf(os.Stderr, "  ./gradlew build 2>&1 | %s\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "\nOPTIONS:\n")
        flag.PrintDefaults()
    }
    var cfg = Config{}
    cfg.Logfile = flag.String("l", "build.log", "Path to save complete build log in")
    cfg.NoLogfile = flag.Bool("N", false, "Do not save a copy of the complete build log")
    cfg.NoWarnings = flag.Bool("W", false, "Ignore warnings")
    flag.Parse()

    tasks, issues, timeTaken := parseBuildLog(&cfg)

    for _,issue := range issues {
        if issue.IsError {
            fmt.Printf("\033[91m\033[0m  %s: %s\n", issue.Location, issue.Desc)
        } else {
            fmt.Printf("\033[93m\033[0m  %s: %s\n", issue.Location, issue.Desc)
        }
    }

    if buildOk(tasks, issues) {
        fmt.Printf("\033[92mBUILD SUCCESSFUL\033[0m in %ds\n", timeTaken)
        os.Exit(0)

    } else {
        fmt.Printf("\033[91mBUILD FAILED\033[0m in %ds\n", timeTaken)
        if len(tasks) == 0 {
            println("No tasks completed")
        }
        if ! *cfg.NoLogfile {
            fmt.Printf("See %s for more information\n", *cfg.Logfile)
        }
        os.Exit(1)
    }
}
