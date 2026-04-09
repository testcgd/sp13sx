package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"sp13sx/internal/app"
	"sp13sx/internal/config"
	"sp13sx/internal/llm/mock"
)

func main() {
	args, configPath, err := extractGlobalConfigArg(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "sp13sx: %v\n", err)
		os.Exit(1)
	}
	if configPath != "" {
		os.Setenv("SP13SX_CONFIG", configPath)
	}

	if len(args) < 1 {
		runMain()
		return
	}

	switch args[0] {
	case "test":
		runTest(args[1:])
	case "record":
		runRecord(args[1:])
	case "validate":
		runValidate(args[1:])
	case "help", "-h", "--help":
		printUsage()
	default:
		runMain()
	}
}

func printUsage() {
	fmt.Print(`sp13sx - 终端编程助手

用法:
	sp13sx [-f 配置文件]              启动交互式 TUI
	sp13sx [-f 配置文件] test         使用 Mock 模式测试 TUI
	sp13sx [-f 配置文件] record       录制真实 LLM 交互
	sp13sx [-f 配置文件] validate     验证场景脚本文件

全局选项:
	-f, --config        指定配置文件路径

环境变量:
	SP13SX_CONFIG       配置文件路径
  SP13SX_TEST_MODE    测试模式 (live|mock|playback|record)
  SP13SX_SCENARIO     场景脚本名称
  SP13SX_RECORDING    录制文件路径 (playback 模式)
  SP13SX_RECORDING_OUTPUT  录制输出路径 (record 模式)
  SP13SX_SMOKE_TEST   设置为 1 运行真实环境冒烟测试

示例:
	# 正常模式
	sp13sx -f ./config.local.yml

	# Mock 模式测试
	sp13sx -f ./config.local.yml test -scenario basic_chat
	./scripts/test-mock.sh basic_chat

  # 录制真实交互
  SP13SX_TEST_MODE=record ./scripts/dev.sh
  ./scripts/test-record.sh

	# 回放录制
	SP13SX_TEST_MODE=playback SP13SX_RECORDING=./test/recordings/demo.jsonl ./scripts/dev.sh
`)
}

func extractGlobalConfigArg(args []string) ([]string, string, error) {
	filtered := make([]string, 0, len(args))
	var configPath string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-f" || arg == "--config":
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("flag needs an argument: %s", arg)
			}
			configPath = args[i+1]
			i++
		case strings.HasPrefix(arg, "-f="):
			configPath = strings.TrimPrefix(arg, "-f=")
		case strings.HasPrefix(arg, "--config="):
			configPath = strings.TrimPrefix(arg, "--config=")
		default:
			filtered = append(filtered, arg)
		}
	}

	if configPath == "" {
		return filtered, "", nil
	}
	return filtered, configPath, nil
}

func runMain() {
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "sp13sx: %v\n", err)
		os.Exit(1)
	}
}

func runTest(args []string) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	scenario := fs.String("scenario", "", "场景脚本名称")
	interactive := fs.Bool("interactive", false, "交互式模式")
	list := fs.Bool("list", false, "列出可用场景")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `用法: sp13sx test [选项]

选项:
`)
		fs.PrintDefaults()
		fmt.Fprint(os.Stderr, `
示例:
  sp13sx test -scenario basic_chat
  sp13sx test -list
  sp13sx test -scenario basic_chat -interactive
`)
	}

	fs.Parse(args)

	if *list {
		listScenarios()
		return
	}

	if *scenario != "" {
		os.Setenv("SP13SX_TEST_MODE", "mock")
		os.Setenv("SP13SX_SCENARIO", *scenario)
	}

	if *interactive {
		runMain()
		return
	}

	if *scenario == "" {
		fmt.Fprintln(os.Stderr, "错误: 非交互模式必须指定 -scenario 参数")
		fs.Usage()
		os.Exit(1)
	}

	// 非交互式模式：运行测试
	fmt.Printf("运行场景测试: %s\n", *scenario)
	runMain()
}

func runRecord(args []string) {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	output := fs.String("output", "", "录制输出路径")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `用法: sp13sx record [选项]

选项:
`)
		fs.PrintDefaults()
		fmt.Fprint(os.Stderr, `
示例:
  sp13sx record -output ./test/recordings/demo.jsonl
`)
	}

	fs.Parse(args)

	os.Setenv("SP13SX_TEST_MODE", "record")
	if *output != "" {
		os.Setenv("SP13SX_RECORDING_OUTPUT", *output)
	}

	runMain()
}

func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	scenario := fs.String("scenario", "", "场景脚本文件路径")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `用法: sp13sx validate [选项]

选项:
`)
		fs.PrintDefaults()
		fmt.Fprint(os.Stderr, `
示例:
  sp13sx validate -scenario ./test/scenarios/basic_chat.yaml
`)
	}

	fs.Parse(args)

	if *scenario == "" {
		fmt.Fprintln(os.Stderr, "错误: 必须指定 -scenario 参数")
		fs.Usage()
		os.Exit(1)
	}

	if err := validateScenario(*scenario); err != nil {
		fmt.Fprintf(os.Stderr, "场景验证失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("场景验证通过:", *scenario)
}

func loadConfig() (config.Config, error) {
	cfgPath := resolveConfigPath()
	return config.Load(cfgPath)
}

func listScenarios() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	scenarioDir := cfg.Test.ScenarioDir
	if scenarioDir == "" {
		scenarioDir = "./test/scenarios"
	}

	fmt.Println("可用场景脚本:")
	fmt.Printf("  目录: %s\n\n", scenarioDir)

	entries, err := os.ReadDir(scenarioDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取场景目录失败: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() && (len(entry.Name()) > 5 && entry.Name()[len(entry.Name())-5:] == ".yaml") {
			name := entry.Name()[:len(entry.Name())-5]
			fmt.Printf("  - %s\n", name)
		}
	}
}

func validateScenario(path string) error {
	scenario, err := mock.LoadScenario(path)
	if err != nil {
		return err
	}
	return scenario.Validate()
}

func resolveConfigPath() string {
	if path := os.Getenv("SP13SX_CONFIG"); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "examples/config.example.yml"
	}
	return home + "/.sp13sx/config.yml"
}
