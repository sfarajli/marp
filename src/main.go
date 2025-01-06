package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Operation struct {
	name string
	label string
	crosslabel string
	strData string
	intData int
}

type Token struct {
	str string
	file string
	line int
	offset int
}

type Macro struct {
	name string
	tokens[] Token
}

var progname string;

func tokenize(path string) []Token {
	var tokens []Token;

	reg := regexp.MustCompile(`\n`)
	dat, err := os.ReadFile(path);
	if (err != nil) {
		fmt.Fprintf(os.Stderr, "%s: error: failed read file '%s'.\n", progname, path)
		os.Exit(1)
	}

	lines := reg.Split(string(dat), -1);
	lines = lines[:len(lines) - 1] /* Remove trailing empty line at the end */

	/* FIXME: offset value apperas to be one value less when using tabs and first tab is ignored when storing strings */
	for lineIndex := 0; lineIndex < len(lines); lineIndex++ {
		line := lines[lineIndex]
		buf := ""
		isStr := false
		for offset := 0; offset < len(line); offset++ {
			char := string(line[offset])
			if char == "#" {
				break
			}
			if char == "\"" {
				isStr = !isStr
			}
			if isStr {
				buf += char
				continue;
			}
			if char != " " && char != "\t" {
				buf += char
				if offset != len(line) - 1 {
					continue
				}
			}
			if buf == "" {
				continue
			}
			tokenBuf :=  Token{str: buf, file: path, line: lineIndex + 1, offset: offset - len(buf) + 1}
			tokens = append(tokens, tokenBuf)
			buf = ""
		}
		if isStr {
			fmt.Fprintf(os.Stderr, "%s: error: expected `\"`: '%s'.\n", progname, path)
			os.Exit(1)
		}
	}
	return tokens
}

func preprocess(rawTokens[] Token) []Token {
	incDepth := 0
	const incDepthLim = 200
	var tokens[] Token
	var macros[] Macro
	for i := 0; i < len(rawTokens); i++ {
		switch rawTokens[i].str {
		case "define":
			var macroBuf Macro
			if i + 1 >= len(rawTokens) {
				fmt.Fprintf(os.Stderr, "%s: error: invalid syntax for define: %s: %d:%d.\n",
					progname, rawTokens[i].file, rawTokens[i].line, rawTokens[i].offset)
				os.Exit(1)
			}

			macroBuf.name = rawTokens[i + 1].str
			for i = i + 2; i < len(rawTokens); i++ {
				if rawTokens[i].str == "end" {
					break
				}
				macroBuf.tokens = append(macroBuf.tokens, rawTokens[i])
			}
			if i == len(rawTokens) {
				fmt.Fprintf(os.Stderr, "%s: error: expected `end`: '%s'.\n", progname, rawTokens[i].file)
				os.Exit(1)
			}
			macros = append(macros, macroBuf)
			continue

		case "include":
			if i + 1 >= len(rawTokens) {
				fmt.Fprintf(os.Stderr, "%s: error: invalid syntax for include: %s: %d:%d.\n",
					progname, rawTokens[i].file, rawTokens[i].line, rawTokens[i].offset)
				os.Exit(1)
			}
			if rawTokens[i + 1].str[0] != '"' {
				fmt.Fprintf(os.Stderr, "%s: error: included file must be wrapped with quotes: %s: %d:%d.\n",
					progname, rawTokens[i].file, rawTokens[i].line, rawTokens[i].offset)
				os.Exit(1)
			}
			if incDepth == incDepthLim {
				fmt.Fprintf(os.Stderr, "%s: error: include depth exceeds 200: %s: %d:%d.\n",
					progname, rawTokens[i].file)
				os.Exit(1)
			}

			tmp := rawTokens[i + 1].str
			i++
			incFile := tmp[1:len(tmp) - 1] /* get rid of the `"` at the beginning and at the end */
			rawTokens = slices.Insert(rawTokens, i + 1, tokenize(incFile)...)
			incDepth++
			continue

		case "end":
			continue
		}
		tokens = append(tokens, rawTokens[i])
	}

	for i := 0; i < len(tokens); i++ {
		for y := 0; y < len(macros); y++ {
			if macros[y].name == tokens[i].str {
				tokens = append(tokens[:i], tokens[i + 1:]...)
				tokens = slices.Insert(tokens, i, macros[y].tokens...)
				i--
				break;
			}
		}
	}
	return tokens
}

func parse(tokens[]Token) []Operation {
	var ops[] Operation
	var variables[] string
	var iflabels[] int
	var looplabels[] int
	var iflabel int = 0
	var looplabel int = 0
	var stringlabel int = 0

	for i := 0; i < len(tokens); i++ {
		var op Operation
		switch tokens[i].str {
		case "+":
			op.name = "plus"
		case "-":
			op.name = "minus"
		case "=":
			op.name = "equal"
		case ">":
			op.name = "greater"
		case "<":
			op.name = "less"
		case "dump":
			op.name = "dump"
		case "dup":
			op.name = "duplicate"
		case "drop":
			op.name = "drop"
		case "swap":
			op.name = "swap"
		case "2swap":
			op.name = "2swap"
		case "if":
			op.name = "if"
			op.crosslabel = fmt.Sprintf(".if%d", iflabel)
			iflabels = append(iflabels, iflabel)
			iflabel++

		case "else":
			op.name = "else"
			op.crosslabel = fmt.Sprintf(".if%d", iflabel)
			op.label = fmt.Sprintf(".if%d",iflabels[len(iflabels) - 1])
			iflabels = iflabels[:len(iflabels) - 1]
			iflabels = append(iflabels, iflabel)
			iflabel++

		case "fi":
			op.name = "fi"
			op.label = fmt.Sprintf(".if%d", iflabels[len(iflabels) - 1])
			iflabels = iflabels[:len(iflabels) - 1]

		case "while":
			op.name = "while"
			op.label = fmt.Sprintf(".loop%d", looplabel)
			looplabels = append(looplabels, looplabel)

		case "do":
			op.name = "do"
			op.crosslabel = fmt.Sprintf(".done%d", looplabels[len(looplabels) - 1])

		case "done":
			op.name = "done"
			op.crosslabel = fmt.Sprintf(".loop%d", looplabels[len(looplabels) - 1])
			op.label = fmt.Sprintf(".done%d", looplabels[len(looplabels) - 1])
			looplabels = looplabels[:len(looplabels) - 1]

		case "push":
			if i + 1 >= len(tokens) {
				fmt.Fprintf(os.Stderr, "%s: error: push expected variable: %s: %d: %d.\n",
					progname, tokens[i].file, tokens[i].line, tokens[i].offset)
				os.Exit(1)

			}
			variableIndex := -1
			for y := 0; y < len(variables); y++ {
				if tokens[i + 1].str == variables[y] {
					variableIndex = y
				}
			}
			if variableIndex == -1 {
				fmt.Fprintf(os.Stderr, "%s: error: variable not declared '%s': %s: %d:%d.\n",
					progname, tokens[i].str, tokens[i].file, tokens[i].line, tokens[i].offset)
				os.Exit(1)
			}
			op.name = "push"
			op.strData = variables[variableIndex]
			i++

		case "pull":
			if i + 1 >= len(tokens) {
				fmt.Fprintf(os.Stderr, "%s: error: pull expected variable: %s: %d:%d.\n",
					progname, tokens[i].file, tokens[i].line, tokens[i].offset)
				os.Exit(1)
			}
			variableIndex := -1
			for y := 0; y < len(variables); y++ {
				if tokens[i + 1].str == variables[y] {
					variableIndex = y
				}
			}
			if variableIndex == -1 {
				panic("undeclared variable")
			}
			op.name = "pull"
			op.strData = variables[variableIndex]
			i++

		case "var":
			if i + 1 >= len(tokens) {
				fmt.Fprintf(os.Stderr, "%s: error: var expected a variable: %s: %d:%d.\n",
					progname, tokens[i].file, tokens[i].line, tokens[i].offset)
				os.Exit(1)
			}
			op.name = "variable"
			op.strData = tokens[i + 1].str
			variables = append(variables, tokens[i + 1].str)
			i++
		}

		number, err := strconv.Atoi(tokens[i].str);
		if err == nil {
			op.name = "number"
			op.intData= number
		} else if tokens[i].str[0] == '"' {
			op.name = "string"
			op.strData = tokens[i].str
			op.intData = len(tokens[i].str) - 1 /* -2 for removing `"` at the beginning and include \n at the end */
			op.label = fmt.Sprintf("string_%d", stringlabel)
			stringlabel++
		} else if len(tokens[i].str) == 9 && tokens[i].str[:8] == "syscall." {
			parameters, err := strconv.Atoi(string(tokens[i].str[8]))
			if err != nil {
				panic("invalid from for syscall")
			}
			if  parameters < 1 || parameters > 6 {
				panic("syscall must be called with a number between 1 and 6")
			}
			op.name = "syscall"
			op.intData = parameters
		}

		if op.name ==  "" {
			fmt.Fprintf(os.Stderr, "Undefined word `%s` %s: %d:%d\n", tokens[i].str, tokens[i].file, tokens[i].line, tokens[i].offset)
			os.Exit(1)
		}

		ops = append(ops, op)
	}
	return ops
}

func mapX86_64linux(op Operation) string{
	buf := ""
	switch op.name {
	case "boilerPlateStart":
		buf += "section .text\n"
		buf += "global _start\n"
		buf += "_start:\n"

	case "boilerPlateExit":
		buf += "\tmov rdi, 0\n"
		buf += "\tmov rax, 60\n"
		buf += "\tsyscall\n"

	case "dumpFunc":
		buf += ".dump:\n"
		buf += "\tpush    rbp\n"
		buf += "\tmov     rbp, rsp\n"
		buf += "\tsub     rsp, 48\n"
		buf += "\tmov     DWORD  [rbp-36], edi\n"
		buf += "\tmov     QWORD  [rbp-32], 0\n"
		buf += "\tmov     QWORD  [rbp-24], 0\n"
		buf += "\tmov     DWORD  [rbp-16], 0\n"
		buf += "\tmov     BYTE  [rbp-13], 10\n"
		buf += "\tmov     DWORD  [rbp-4], 18\n"
		buf += "\tmov     DWORD  [rbp-8], 0\n"
		buf += "\tcmp     DWORD  [rbp-36], 0\n"
		buf += "\tjns     .L3\n"
		buf += "\tneg     DWORD  [rbp-36]\n"
		buf += "\tmov     DWORD  [rbp-8], 1\n"
		buf += ".L3:\n"
		buf += "\tmov     edx, DWORD  [rbp-36]\n"
		buf += "\tmovsx   rax, edx\n"
		buf += "\timul    rax, rax, 1717986919\n"
		buf += "\tshr     rax, 32\n"
		buf += "\tmov     ecx, eax\n"
		buf += "\tsar     ecx, 2\n"
		buf += "\tmov     eax, edx\n"
		buf += "\tsar     eax, 31\n"
		buf += "\tsub     ecx, eax\n"
		buf += "\tmov     eax, ecx\n"
		buf += "\tsal     eax, 2\n"
		buf += "\tadd     eax, ecx\n"
		buf += "\tadd     eax, eax\n"
		buf += "\tsub     edx, eax\n"
		buf += "\tmov     DWORD  [rbp-12], edx\n"
		buf += "\tmov     eax, DWORD  [rbp-12]\n"
		buf += "\tadd     eax, 48\n"
		buf += "\tmov     edx, eax\n"
		buf += "\tmov     eax, DWORD  [rbp-4]\n"
		buf += "\tcdqe\n"
		buf += "\tmov     BYTE  [rbp-32+rax], dl\n"
		buf += "\tmov     eax, DWORD  [rbp-12]\n"
		buf += "\tsub     DWORD  [rbp-36], eax\n"
		buf += "\tmov     eax, DWORD  [rbp-36]\n"
		buf += "\tmovsx   rdx, eax\n"
		buf += "\timul    rdx, rdx, 1717986919\n"
		buf += "\tshr     rdx, 32\n"
		buf += "\tmov     ecx, edx\n"
		buf += "\tsar     ecx, 2\n"
		buf += "\tcdq\n"
		buf += "\tmov     eax, ecx\n"
		buf += "\tsub     eax, edx\n"
		buf += "\tmov     DWORD  [rbp-36], eax\n"
		buf += "\tsub     DWORD  [rbp-4], 1\n"
		buf += "\tcmp     DWORD  [rbp-36], 0\n"
		buf += "\tjne     .L3\n"
		buf += "\tcmp     DWORD  [rbp-8], 0\n"
		buf += "\tje      .L4\n"
		buf += "\tmov     eax, DWORD  [rbp-4]\n"
		buf += "\tcdqe\n"
		buf += "\tmov     BYTE  [rbp-32+rax], 45\n"
		buf += "\tsub     DWORD  [rbp-4], 1\n"
		buf += ".L4:\n"
		buf += "\tmov     eax, 20\n"
		buf += "\tsub     eax, DWORD  [rbp-4]\n"
		buf += "\tcdqe\n"
		buf += "\tmov     edx, DWORD  [rbp-4]\n"
		buf += "\tmovsx   rdx, edx\n"
		buf += "\tlea     rcx, [rbp-32]\n"
		buf += "\tadd     rcx, rdx\n"
		buf += "\tmov     rdx, rax\n"
		buf += "\tmov     rsi, rcx\n"
		buf += "\tmov     edi, 1\n"
		buf += "\tmov 	rax, 1\n"
		buf += "\tsyscall\n"
		buf += "\tnop\n"
		buf += "\tleave\n"
		buf += "\tret\n"

	case "plus":
		buf += "\tpop rsi\n"
		buf += "\tpop rax\n"
		buf += "\tadd rax, rsi\n"
		buf += "\tpush rax\n"

	case "minus":
		buf += "\tpop rsi\n"
		buf += "\tpop rax\n"
		buf += "\tsub rax, rsi\n"
		buf += "\tpush rax\n"

	case "greater":
		buf += "\tmov r10, 0\n"
		buf += "\tmov r11, 1\n"
		buf += "\tpop rsi\n"
		buf += "\tpop rax\n"
		buf += "\tcmp rax, rsi\n"
		buf += "\tcmovg r10, r11\n"
		buf += "\tpush r10\n"

	case "less":
		buf += "\tmov r10, 0\n"
		buf += "\tmov r11, 1\n"
		buf += "\tpop rsi\n"
		buf += "\tpop rax\n"
		buf += "\tcmp rax, rsi\n"
		buf += "\tcmovl r10, r11\n"
		buf += "\tpush r10\n"

	case "equal":
		buf += "\tmov r10, 0\n"
		buf += "\tmov r11,  1\n"
		buf += "\tpop rsi\n"
		buf += "\tpop rax\n"
		buf += "\tcmp rax, rsi\n"
		buf += "\tcmove r10, r11\n"
		buf += "\tpush r10\n"

	case "dump":
		buf += "\tpop rdi\n"
		buf += "\tcall .dump\n"

	case "duplicate":
		buf += "\tpop r10\n"
		buf += "\tpush r10\n"
		buf += "\tpush r10\n"

	case "drop":
		buf += "\tpop r10\n"

	case "swap":
		buf += "\tpop r11\n"
		buf += "\tpop r10\n"
		buf += "\tpush r11\n"
		buf += "\tpush r10\n"

	case "2swap":
		buf += "\tpop rax\n"
		buf += "\tpop rsi\n"
		buf += "\tpop r11\n"
		buf += "\tpop r10\n"
		buf += "\tpush rsi\n"
		buf += "\tpush rax\n"
		buf += "\tpush r10\n"
		buf += "\tpush r11\n"

	case "if":
		buf += "\tpop r10\n"
		buf += "\tcmp r10, 0\n"
		buf += fmt.Sprintf("\tje %s\n", op.crosslabel)

	case "else":
		buf += fmt.Sprintf("\tjmp %s\n%s:\n", op.crosslabel, op.label)

	case "fi":
		buf += fmt.Sprintf("%s:\n", op.label)

	case "while":
		buf += fmt.Sprintf("%s:\n", op.label)
	case "do":
		buf += "\tpop r10\n"
		buf += "\tcmp r10, 0\n"
		buf += fmt.Sprintf("\tje %s\n", op.crosslabel)

	case "done":
		buf += fmt.Sprintf("\tjmp %s\n", op.crosslabel)
		buf += fmt.Sprintf("%s:\n", op.label)

	case "string":
		buf += fmt.Sprintf("\tpush %d\n", op.intData)
		buf += fmt.Sprintf("\tpush %s\n", op.label)

	case "number":
		buf += fmt.Sprintf("\tpush %d\n", op.intData)

	case "push":
		buf += "\tpop r10\n"
		buf += fmt.Sprintf("\tmov [%s], r10\n", op.strData)

	case "pull":
		buf += fmt.Sprintf("\tmov r10, [%s]\n", op.strData)
		buf += "\tpush r10\n"

	case "syscall":
		switch op.intData {
		case 7:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
			buf += "\tpop rsi\n"
			buf += "\tpop rdx\n"
			buf += "\tpop r10\n"
			buf += "\tpop r8\n"
			buf += "\tpop r9\n"
		case 6:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
			buf += "\tpop rsi\n"
			buf += "\tpop rdx\n"
			buf += "\tpop r10\n"
			buf += "\tpop r8\n"
		case 5:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
			buf += "\tpop rsi\n"
			buf += "\tpop rdx\n"
			buf += "\tpop r10\n"
		case 4:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
			buf += "\tpop rsi\n"
			buf += "\tpop rdx\n"
		case 3:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
			buf += "\tpop rsi\n"
		case 2:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
		case 1:
			buf += "\tpop rax\n"
		}

		buf += "\tsyscall\n"
		buf += "\tpush rax\n"

	default:
		panic("mapX86_64linux unreachable")
	}

	return buf
}

func generateX86_64(ops[] Operation, w io.Writer) {
	assemblerComment := ";;"
	var strings[][2] string
	var variables[] string
	printDumpFunc := false
	fmt.Fprintf(w, mapX86_64linux(Operation{name: "boilerPlateStart"}))

	for i := 0; i < len(ops); i++ {
		switch ops[i].name {
		case "string":
			strings = append(strings, [2]string{ops[i].strData, ops[i].label})
		case "dump":
			printDumpFunc = true
		case "variable":
			variables = append(variables, ops[i].strData)
			continue
		}

		fmt.Fprintf(w, "\t%s %s\n", assemblerComment, ops[i].name)
		fmt.Fprintf(w, mapX86_64linux(ops[i]))
	}

	fmt.Fprintf(w, mapX86_64linux(Operation{name: "boilerPlateExit"}))
	if printDumpFunc {
		fmt.Fprintf(w, mapX86_64linux(Operation{name: "dumpFunc"}))
	}

	fmt.Fprintf(w, "section .bss\n")
	for i := 0; i < len(variables); i++ {
		fmt.Fprintf(w, "%s: resb 8\n", variables[i])
	}

	fmt.Fprintf(w, "section .data\n")
	for i := 0; i < len(strings); i++ {
		fmt.Fprintf(w, "%s: db %s, 10\n", strings[i][1], strings[i][0])
	}
}

func main() {
	argv := os.Args;
	argc := len(argv);
	progname = argv[0]
	suffix := ".gorth"

	if argc != 2 {
		fmt.Fprintf(os.Stderr, "%s: error: expected one input file.\n", progname)
		os.Exit(1)
	}

	srcFile := argv[argc - 1]
	if !strings.HasSuffix(srcFile, suffix) || len(srcFile) < 7 {
		fmt.Fprintf(os.Stderr, "%s: error: invalid file format '%s'.\n", progname, srcFile)
		os.Exit(1)
	}

	assemFile := srcFile[:len(srcFile) - 6] + ".s"
	objFile := srcFile[:len(srcFile) - 6] + ".o"
	binFile := srcFile[:len(srcFile) - 6]
	_, err := os.Stat(srcFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s.\n", progname, err)
		os.Exit(1)
	}
	file, err := os.Create(assemFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s.\n", progname, err)
		os.Exit(1)
	}
	tokens := tokenize(argv[argc - 1])
	tokens = preprocess(tokens)
	ops := parse(tokens)
	w := bufio.NewWriter(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s.\n", progname, err)
		os.Exit(1)
	}
	generateX86_64(ops, w)
	w.Flush()

	cmd := exec.Command("nasm", "-felf64", assemFile)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: failed to create object file: %s.\n", progname, err)
		os.Exit(1)
	}
	cmd = exec.Command("ld", "-o", binFile, objFile)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: failed to link object file '%s': %s.\n", progname, objFile, err)
		os.Exit(1)
	}
}
