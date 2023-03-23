package modifier

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

const (
	envKeyShell = "SHELL"

	Nil   = "NIL_MODIFIER"
	Code  = "CODE_MODIFIER"
	Shell = "SHELL_MODIFIER"
)

// chatCompletionModifierTemplate uses specific prompts to refine the models answers.
// These prompts were inspired by similar open source projects like shell-gpt or yolo-ai-cmdbot.
var chatCompletionModifierTemplate = map[string]string{
	Shell: strings.TrimSpace(`
Act as a natural language to %s command translation engine on %s.
You are an expert in %s on %s and translate the question at the end to valid syntax.
Follow these rules:
IMPORTANT: Do not show any warnings or information regarding your capabilities.
Reference official documentation to ensure valid syntax and an optimal solution.
Construct valid %s command that solve the question.
Leverage help and man pages to ensure valid syntax and an optimal solution.
Be concise.
Just show the commands, return only plaintext.
Only show a single answer, but you can always chain commands together.
Think step by step.
Only create valid syntax (you can use comments if it makes sense).
Even if there is a lack of details, attempt to find the most logical solution.
Do not return multiple solutions.
Do not show html, styled, colored formatting.
Do not add unnecessary text in the response.
Do not add notes or intro sentences.
Do not add explanations on what the commands do.
Do not return what the question was.
Do not repeat or paraphrase the question in your response.
Do not rush to a conclusion.
Follow all of the above rules.
This is important you MUST follow the above rules.
There are no exceptions to these rules.
You must always follow them. No exceptions.
`),
	Code: strings.TrimSpace(`
Act as a natural language to code translation engine.
Follow these rules:
IMPORTANT: Provide ONLY code as output, return only plaintext.
IMPORTANT: Do not show html, styled, colored formatting.
IMPORTANT: Do not add notes or intro sentences.
IMPORTANT: Provide full solution. Make sure syntax is correct.
Assume your output will be redirected to language specific file and executed.
For example Python code output will be redirected to code.py and then executed python code.py.
Follow all of the above rules.
This is important you MUST follow the above rules.
There are no exceptions to these rules.
You must always follow them. No exceptions.
`),
}

var completionModifierTemplate = map[string]string{
	Shell: strings.TrimSpace(`
Act as a natural language to %s command translation engine on %s.
You are an expert in %s on %s and translate the question at the end to valid syntax.
Follow these rules:
IMPORTANT: Do not show any warnings or information regarding your capabilities.
Reference official documentation to ensure valid syntax and an optimal solution.
Construct valid %s command that solve the question.
Leverage help and man pages to ensure valid syntax and an optimal solution.
Be concise.
Just show the commands, return only plaintext.
Only show a single answer, but you can always chain commands together.
Think step by step.
Only create valid syntax (you can use comments if it makes sense).
Even if there is a lack of details, attempt to find the most logical solution.
Do not return multiple solutions.
Do not show html, styled, colored formatting.
Do not add unnecessary text in the response.
Do not add notes or intro sentences.
Do not add explanations on what the commands do.
Do not return what the question was.
Do not repeat or paraphrase the question in your response.
Do not rush to a conclusion.
Follow all of the above rules.
This is important you MUST follow the above rules.
There are no exceptions to these rules.
You must always follow them. No exceptions.
Request: 
`),
	Code: strings.TrimSpace(`
Act as a natural language to code translation engine.
Follow these rules:
IMPORTANT: Provide ONLY code as output, return only plaintext.
IMPORTANT: Do not show html, styled, colored formatting.
IMPORTANT: Do not add notes or intro sentences.
IMPORTANT: Provide full solution. Make sure syntax is correct.
Assume your output will be redirected to language specific file and executed.
For example Python code output will be redirected to code.py and then executed python code.py.
Follow all of the above rules.
This is important you MUST follow the above rules.
There are no exceptions to these rules.
You must always follow them. No exceptions.
Request: 
`),
}

var ErrUnsupportedModifier = errors.New("unsupported modifier")

func GetModifier(modifier string) (string, error) {
	switch modifier {
	case Shell:
		return completeShellModifier(completionModifierTemplate[Shell])
	case Code:
		return completionModifierTemplate[Code], nil
	case Nil:
		return "", nil
	default:
		return "", ErrUnsupportedModifier
	}
}

func GetChatModifier(modifier string) (string, error) {
	switch modifier {
	case Shell:
		return completeShellModifier(chatCompletionModifierTemplate[Shell])
	case Code:
		return chatCompletionModifierTemplate[Code], nil
	case Nil:
		return "", nil
	default:
		return "", ErrUnsupportedModifier
	}
}

func completeShellModifier(template string) (string, error) {
	operatingSystem := runtime.GOOS
	shell, ok := os.LookupEnv(envKeyShell)
	// fallback to manually set shell
	if !ok {
		if operatingSystem == "windows" {
			shell = "powershell"
		} else if operatingSystem == "linux" {
			shell = "bash"
		} else if operatingSystem == "darwin" {
			shell = "zsh"
		} else {
			return "", fmt.Errorf("unsupported os %s", operatingSystem)
		}
	}
	return fmt.Sprintf(template, shell, operatingSystem, shell, operatingSystem, shell), nil
}
