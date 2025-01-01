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

type Word struct {
	typ string
	name string
	tokens[] Token
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
	var externalWords[] Word;

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

		case "const":
			if i + 2 >= len(tokens) {
				/* FIXME: Better error message */
				panic("invalid const usage")
			}
			name := tokens[i + 1].str
			token := tokens[i + 2]
			_, err := strconv.Atoi(token.str)
			/* FIXME: add character as well */
			if err != nil && token.str[0] != '"' {
				/* FIXME: Better error message */
				panic("Can only define strings and numbers with const")
			}
			externalWords = append(externalWords, Word{typ: "const", name: name, tokens: []Token{token}})
			i += 2
			continue
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

		for y := 0; y < len(externalWords); y++ {
			if tokens[i].str != externalWords[y].name {
				continue
			}
			opsBuf := parse(externalWords[y].tokens)
			ops = append(ops, opsBuf...)
			/* FIXME: Find an alternative to goto */
			goto done
		}

		if op.name ==  "" {
			panic ("invalid word")
		}

		ops = append(ops, op)
done:
	}
	return ops
}

/* FIXME: find a better function name */
func X86_64map(op Operation) string{
	switch op.name {
	case "plus":
		return "\tpop rsi\n\tpop rax\n\tadd rax, rsi\n\tpush rax\n"
	case "minus":
		return "\tpop rsi\n\tpop rax\n\tsub rax, rsi\n\tpush rax\n"
	case "greater":
		return "\tmov r10, 0\n\tmov r11, 1\n\tpop rsi\n\tpop rax\n\tcmp rax, rsi\n\tcmovg r10, r11\n\tpush r10\n"
	case "less":
		return "\tmov r10, 0\n\tmov r11, 1\n\tpop rsi\n\tpop rax\n\tcmp rax, rsi\n\tcmovl r10, r11\n\tpush r10\n"
	case "equal":
		return "\tmov r10, 0\n\tmov r11,  1\n\tpop rsi\n\tpop rax\n\tcmp rax, rsi\n\tcmove r10, r11\n\tpush r10\n"
	case "dump":
		return "\tpop rdi\n\tcall .dump\n"
	case "duplicate":
		return "\tpop r10\n\tpush r10\n\tpush r10\n"
	case "drop":
		return "\tpop r10\n"
	case "swap":
		return "\tpop r11\n\tpop r10\n\tpush r11\n\tpush r10\n"
	case "2swap":
		return "\tpop rax\n\tpop rsi\n\tpop r11\n\tpop r10\n\tpush rsi\n\tpush rax\n\tpush r10\n\tpush r11\n"
	case "if":
		return fmt.Sprintf("\tpop r10\n\tcmp r10, 0\n\tje %s\n", op.crosslabel)
	case "else":
		return fmt.Sprintf("\tjmp %s\n%s:\n", op.crosslabel, op.label)
	case "fi":
		return fmt.Sprintf("%s:\n", op.label)
	case "while":
		return fmt.Sprintf("%s:\n", op.label)
	case "do":
		return fmt.Sprintf("\tpop r10\n\tcmp r10, 0\n\tje %s\n", op.crosslabel)
	case "done":
		return fmt.Sprintf("\tjmp %s\n%s:\n", op.crosslabel, op.label)
	case "string":
		return fmt.Sprintf("\tpush %d\n\tpush %s\n", op.intData, op.label)
	case "number":
		return fmt.Sprintf("\tpush %d\n", op.intData)
	case "syscall":
		/* FIXME: make it more consise */
		buf := ""
		switch op.intData {
		case 7:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
			buf += "\tpop rsi\n"
			buf += "\tpop rdx\n"
			buf += "\tpop r10\n"
			buf +="\tpop r8\n"
			buf +="\tpop r9\n"
		case 6:
			buf += "\tpop rax\n"
			buf += "\tpop rdi\n"
			buf += "\tpop rsi\n"
			buf += "\tpop rdx\n"
			buf += "\tpop r10\n"
			buf +="\tpop r8\n"
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
		return buf
	}
	panic("X86_64map unreachable")
}

func compileX86_64(ops[] Operation) {
	var strings[][2] string
	print("section .text")
	print("global _start")
	print("_start:")

	for i := 0; i < len(ops); i++ {
		fmt.Printf("\t;; %s\n", ops[i].name)
		fmt.Printf(X86_64map(ops[i]))
		if (ops[i].name == "string") {
			strings = append(strings, [2]string{ops[i].strData, ops[i].label})
		}
	}

	print("	;; EXIT")
	print("	mov rdi, 0")
	print("	mov rax, 60")
	print("	syscall")

	print(".dump:");
	print("	push    rbp")
	print("	mov     rbp, rsp")
	print("	sub     rsp, 48")
	print("	mov     DWORD  [rbp-36], edi")
	print("	mov     QWORD  [rbp-32], 0")
	print("	mov     QWORD  [rbp-24], 0")
	print("	mov     DWORD  [rbp-16], 0")
	print("	mov     BYTE  [rbp-13], 10")
	print("	mov     DWORD  [rbp-4], 18")
	print("	mov     DWORD  [rbp-8], 0")
	print("	cmp     DWORD  [rbp-36], 0")
	print("	jns     .L3")
	print("	neg     DWORD  [rbp-36]")
	print("	mov     DWORD  [rbp-8], 1")
	print(".L3:")
	print("	mov     edx, DWORD  [rbp-36]")
	print("	movsx   rax, edx")
	print("	imul    rax, rax, 1717986919")
	print("	shr     rax, 32")
	print("	mov     ecx, eax")
	print("	sar     ecx, 2")
	print("	mov     eax, edx")
	print("	sar     eax, 31")
	print("	sub     ecx, eax")
	print("	mov     eax, ecx")
	print("	sal     eax, 2")
	print("	add     eax, ecx")
	print("	add     eax, eax")
	print("	sub     edx, eax")
	print("	mov     DWORD  [rbp-12], edx")
	print("	mov     eax, DWORD  [rbp-12]")
	print("	add     eax, 48")
	print("	mov     edx, eax")
	print("	mov     eax, DWORD  [rbp-4]")
	print("	cdqe")
	print("	mov     BYTE  [rbp-32+rax], dl")
	print("	mov     eax, DWORD  [rbp-12]")
	print("	sub     DWORD  [rbp-36], eax")
	print("	mov     eax, DWORD  [rbp-36]")
	print("	movsx   rdx, eax")
	print("	imul    rdx, rdx, 1717986919")
	print("	shr     rdx, 32")
	print("	mov     ecx, edx")
	print("	sar     ecx, 2")
	print("	cdq")
	print("	mov     eax, ecx")
	print("	sub     eax, edx")
	print("	mov     DWORD  [rbp-36], eax")
	print("	sub     DWORD  [rbp-4], 1")
	print("	cmp     DWORD  [rbp-36], 0")
	print("	jne     .L3")
	print("	cmp     DWORD  [rbp-8], 0")
	print("	je      .L4")
	print("	mov     eax, DWORD  [rbp-4]")
	print("	cdqe")
	print("	mov     BYTE  [rbp-32+rax], 45")
	print("	sub     DWORD  [rbp-4], 1")
	print(".L4:")
	print("	mov     eax, 20")
	print("	sub     eax, DWORD  [rbp-4]")
	print("	cdqe")
	print("	mov     edx, DWORD  [rbp-4]")
	print("	movsx   rdx, edx")
	print("	lea     rcx, [rbp-32]")
	print("	add     rcx, rdx")
	print("	mov     rdx, rax")
	print("	mov     rsi, rcx")
	print("	mov     edi, 1")
	print("	mov 	rax, 1")
	print("	syscall")
	print("	nop")
	print("	leave")
	print("	ret")

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
