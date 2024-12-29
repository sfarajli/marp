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
	data string
}

type Token struct {
	str string
	line int
	offset int
}

/* TODO: add error checking */

func tokenize(path string) []Token {
	var tokens []Token;

	reg := regexp.MustCompile(`\n`)
	dat, err := os.ReadFile(path);
	if (err != nil) {
		panic("Failed to read file");
	}

	lines := reg.Split(string(dat), -1);
	lines = lines[:len(lines) - 1] /* Remove trailing empty line at the end */

	/* FIXME: offset value apperas to be one value less when using tabs */
	for lineIndex := 0; lineIndex < len(lines); lineIndex++ {
		line := lines[lineIndex]
		buf := ""
		for offset := 0; offset < len(line); offset++ {
			char := string(line[offset])
			if char == "#" {
				break
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
	}
	return tokens
}

func parse(tokens[]Token) []Operation {
	var ops[] Operation
	var iflabels[] int
	var iflabel int = 0
	for i := 0; i < len(tokens); i++ {
		var op Operation
		_, err := strconv.Atoi(tokens[i].str);
		if err == nil {
			op.name = "number"
			op.data = tokens[i].str
			/* FIXME: append for ops called at the end */
			ops = append(ops, op)
			continue;
		}
		if tokens[i].str[0] == '"' {
			panic("string parsing is not implemented")
		}

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
		case ".":
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

		case "while": fallthrough
		case "do": fallthrough
		case "done":
			panic("Loops not implemented")
		default:
			panic("invalid word")
		}
		ops = append(ops, op)
	}
	return ops
}

func compileX86_64(ops[] Operation) {
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

		case "rem":
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

		case "while": fallthrough
		case "do": fallthrough
		case "done":
			panic("Unreachable: loops not implemented")

		case "number":
			number, err := strconv.Atoi(ops[i].data)
			if err != nil {
				panic("Unreachable")
			}
			fmt.Printf("	push %d\n", number)

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
