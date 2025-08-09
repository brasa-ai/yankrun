package helpers

import (
    "os"

    "github.com/rs/zerolog"
)

var Log zerolog.Logger

func SetupLogger(level string) {
    var l zerolog.Level
    switch level {
    case "debug":
        l = zerolog.DebugLevel
    case "info":
        l = zerolog.InfoLevel
    case "warn":
        l = zerolog.WarnLevel
    case "error":
        l = zerolog.ErrorLevel
    default:
        l = zerolog.InfoLevel
    }
    zerolog.SetGlobalLevel(l)
}

func init() {
    zerolog.SetGlobalLevel(zerolog.InfoLevel)
    output := zerolog.ConsoleWriter{
        Out:        os.Stderr,
        TimeFormat: "\u001B[90m15:04:05\u001B[0m",
        NoColor:    false,
        PartsOrder: []string{
            zerolog.TimestampFieldName,
            zerolog.LevelFieldName,
            zerolog.MessageFieldName,
        },
        FormatLevel: func(i interface{}) string {
            level := "????"
            if s, ok := i.(string); ok {
                switch s {
                case "debug":
                    level = "\u001B[36mDEBUG\u001B[0m"
                case "info":
                    level = "\u001B[32mINFO \u001B[0m"
                case "warn":
                    level = "\u001B[33mWARN \u001B[0m"
                case "error":
                    level = "\u001B[31mERROR\u001B[0m"
                case "fatal":
                    level = "\u001B[31mFATAL\u001B[0m"
                }
            }
            return level
        },
        FormatMessage: func(i interface{}) string {
            if i == nil {
                return ""
            }
            return "  " + i.(string)
        },
    }
    Log = zerolog.New(output).With().Timestamp().Logger()
}


