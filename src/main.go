package main

import (
	"fmt"
	"regexp"
	"strconv"
	"os"
)

var print = fmt.Println

type Operation struct {
	name string
	label string
	crosslabel string
	strData string
	intData int
}

type Token struct {
	str string
	line int
	offset int
}

/* TODO: add validate function for error checking */

func tokenize(path string) []Token {
	var tokens []Token;

	reg := regexp.MustCompile(`\n`)
	dat, err := os.ReadFile(path);
	if (err != nil) {
		panic("Failed to read file");
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
			tokenBuf :=  Token{str: buf, line: lineIndex + 1, offset: offset - len(buf) + 1}
			tokens = append(tokens, tokenBuf)
			buf = ""
		}
		if isStr {
			panic("invalid syntax")
		}
	}
	return tokens
}

func parse(tokens[]Token) []Operation {
	var ops[] Operation
	var iflabels[] int
	var iflabel int = 0
	var looplabels[] int
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
		case "rem":
			op.name = "remove"
		case "swap":
			op.name = "swap"
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

		default:
			number, err := strconv.Atoi(tokens[i].str);
			if err == nil {
				op.name = "number"
				op.intData= number
			} else if tokens[i].str[0] == '"' {
				op.name = "string"
				op.strData = tokens[i].str
				op.intData = len(tokens[i].str)
				op.label = fmt.Sprintf("string_%d", stringlabel)
				stringlabel++
			} else {
				panic ("invalid word")
			}
		}
		ops = append(ops, op)
	}
	return ops
}

func compileX86_64(ops[] Operation) {
	var strings[][2] string
	print("section .text")
	print("global _start")
	print("_start:")

	for i := 0; i < len(ops); i++ {
		fmt.Printf("		;; %s\n", ops[i].name)
		switch ops[i].name {
		case "plus":
			print("		pop rsi")
			print("		pop rax")
			print("		add rax, rsi")
			print("		push rax")

		case "minus":
			print("		pop rsi")
			print("		pop rax")
			print("		sub rax, rsi")
			print("		push rax")

		case "greater":
			print("		mov r10, 0")
			print("		mov r11, 1")
			print("		pop rsi")
			print("		pop rax")
			print("		cmp rax, rsi")
			print("		cmovg r10, r11")
			print("		push r10")

		case "less":
			print("		mov r10, 0")
			print("		mov r11, 1")
			print("		pop rsi")
			print("		pop rax")
			print("		cmp rax, rsi")
			print("		cmovl r10, r11")
			print("		push r10")

		case "equal":
			print("		mov r10, 0")
			print("		mov r11,  1")
			print("		pop rsi")
			print("		pop rax")
			print("		cmp rax, rsi")
			print("		cmove r10, r11")
			print("		push r10")

		case "dump":
			print("		pop rdi")
			print("		call .dump")

		case "duplicate":
			print("		pop r10")
			print("		push r10")
			print("		push r10")

		case "remove":
			print("		pop r10")

		case "swap":
			print("		pop r11")
			print("		pop r10")
			print("		push r11")
			print("		push r10")

		case "if":
			print("		pop r10")
			print("		cmp r10, 0")
			fmt.Printf("		je %s\n", ops[i].crosslabel)

		case "else":
			fmt.Printf("		jmp %s\n", ops[i].crosslabel)
			fmt.Printf("%s:\n", ops[i].label)

		case "fi":
			fmt.Printf("%s:\n", ops[i].label)

		case "while":
			fmt.Printf("%s:\n", ops[i].label)

		case "do":
			print("		pop r10")
			print("		cmp r10, 0")
			fmt.Printf("		je %s\n", ops[i].crosslabel)

		case "done":
			fmt.Printf("		jmp %s\n", ops[i].crosslabel)
			fmt.Printf("%s:\n", ops[i].label)

		case "string":
			fmt.Printf("		push %d\n", ops[i].intData)
			fmt.Printf("		push %s\n", ops[i].label)
			strings = append(strings, [2]string {ops[i].strData, ops[i].label})

		case "number":
			fmt.Printf("		push %d\n", ops[i].intData)

		default:
			panic("Unreachable")
		}
	}

	print("		;; EXIT")
	print("		mov rdi, 0")
	print("		mov rax, 60")
	print("		syscall")


	print(".dump:");
	print("		push    rbp")
	print("		mov     rbp, rsp")
	print("		sub     rsp, 48")
	print("		mov     DWORD  [rbp-36], edi")
	print("		mov     QWORD  [rbp-32], 0")
	print("		mov     QWORD  [rbp-24], 0")
	print("		mov     DWORD  [rbp-16], 0")
	print("		mov     BYTE  [rbp-13], 10")
	print("		mov     DWORD  [rbp-4], 18")
	print("		mov     DWORD  [rbp-8], 0")
	print("		cmp     DWORD  [rbp-36], 0")
	print("		jns     .L3")
	print("		neg     DWORD  [rbp-36]")
	print("		mov     DWORD  [rbp-8], 1")
	print(".L3:")
	print("		mov     edx, DWORD  [rbp-36]")
	print("		movsx   rax, edx")
	print("		imul    rax, rax, 1717986919")
	print("		shr     rax, 32")
	print("		mov     ecx, eax")
	print("		sar     ecx, 2")
	print("		mov     eax, edx")
	print("		sar     eax, 31")
	print("		sub     ecx, eax")
	print("		mov     eax, ecx")
	print("		sal     eax, 2")
	print("		add     eax, ecx")
	print("		add     eax, eax")
	print("		sub     edx, eax")
	print("		mov     DWORD  [rbp-12], edx")
	print("		mov     eax, DWORD  [rbp-12]")
	print("		add     eax, 48")
	print("		mov     edx, eax")
	print("		mov     eax, DWORD  [rbp-4]")
	print("		cdqe")
	print("		mov     BYTE  [rbp-32+rax], dl")
	print("		mov     eax, DWORD  [rbp-12]")
	print("		sub     DWORD  [rbp-36], eax")
	print("		mov     eax, DWORD  [rbp-36]")
	print("		movsx   rdx, eax")
	print("		imul    rdx, rdx, 1717986919")
	print("		shr     rdx, 32")
	print("		mov     ecx, edx")
	print("		sar     ecx, 2")
	print("		cdq")
	print("		mov     eax, ecx")
	print("		sub     eax, edx")
	print("		mov     DWORD  [rbp-36], eax")
	print("		sub     DWORD  [rbp-4], 1")
	print("		cmp     DWORD  [rbp-36], 0")
	print("		jne     .L3")
	print("		cmp     DWORD  [rbp-8], 0")
	print("		je      .L4")
	print("		mov     eax, DWORD  [rbp-4]")
	print("		cdqe")
	print("		mov     BYTE  [rbp-32+rax], 45")
	print("		sub     DWORD  [rbp-4], 1")
	print(".L4:")
	print("		mov     eax, 20")
	print("		sub     eax, DWORD  [rbp-4]")
	print("		cdqe")
	print("		mov     edx, DWORD  [rbp-4]")
	print("		movsx   rdx, edx")
	print("		lea     rcx, [rbp-32]")
	print("		add     rcx, rdx")
	print("		mov     rdx, rax")
	print("		mov     rsi, rcx")
	print("		mov     edi, 1")
	print("		mov 	rax, 1")
	print("		syscall")
	print("		nop")
	print("		leave")
	print("		ret")

	print("section .data")
	for i := 0; i < len(strings); i++ {
		fmt.Printf("%s: db %s, 10\n", strings[i][1], strings[i][0])
	}
}

func main() {
	argv := os.Args;
	argc := len(argv);

	if (argc != 2) {
		panic("Invaid usage");
	}

	tokens := tokenize(argv[argc - 1])
	ops := parse(tokens)
	compileX86_64(ops)
}
