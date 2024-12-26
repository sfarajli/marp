package main

import (
	"fmt"
	"regexp"
	"strconv"
	"os"
)

var print = fmt.Println

func readfile(path string) []string {

	var raw []string;
	reg := regexp.MustCompile(`\s`)

	dat, err := os.ReadFile(path);
	if (err != nil) {
		panic("failed to read file");
	}

	tmp := reg.Split(string(dat), -1);

	for i := range tmp{
		if tmp[i] != "" && tmp[i][0] != '#' {
			raw = append(raw, tmp[i]);
		}
	};

	return raw;
}

func compile(raw[]string) {
	var iflabels[] int;
	var looplabels[] int;
	print("section .text")
	print("global _start")
	print("_start:")

	for i := 0; i < len(raw); i++ {
		number, err := strconv.Atoi(raw[i]);
		if err == nil {
			print("		;; PUSHING TO STACK")
			print("		push", number);
			continue;
		}

		switch raw[i] {
		case "+":
			print("		;; PLUS")
			print("		pop rsi")
			print("		pop rax")
			print("		add rax, rsi")
			print("		push rax")
		case "-":
			print("		;; MINUS")
			print("		pop rsi")
			print("		pop rax")
			print("		sub rax, rsi")
			print("		push rax")
		case ".":
			print("		;; DUMP")
			print("		pop rdi")
			print("		call .dump")
		case "=":
			print("		;; EQUAL")
			print("		mov r10, 0")
			print("		mov r11,  1")
			print("		pop rsi")
			print("		pop rax")
			print("		cmp rsi, rax")
			print("		cmove r10, r11")
			print("		push r10")
		case "if":
			print("		;; IF")
			print("		pop r10")
			print("		cmp r10, 0")
			fmt.Printf("		je .if%d\n", i)
			iflabels = append(iflabels, i)

		case "else":
			print("		;; ELSE")
			fmt.Printf("		jmp .if%d\n", i)
			fmt.Printf(".if%d:\n", iflabels[len(iflabels) - 1])
			iflabels = iflabels[:len(iflabels) - 1]
			iflabels = append(iflabels, i)

		case "endif":
			print("		;; ENDIF")
			fmt.Printf(".if%d:\n", iflabels[len(iflabels) - 1])
			iflabels = iflabels[:len(iflabels) - 1]

		case "while":
			print("		;; WHILE")
			fmt.Printf(".loop%d:\n", i);
			looplabels = append(looplabels, i)

		case "do":
			print("		;; DO")
			print("		pop r10")
			print("		cmp r10, 0")
			fmt.Printf("		je .endloop%d\n", looplabels[len(looplabels) - 1])

		case "endloop":
			print("		;; ENDLOOP")
			fmt.Printf("		jmp .loop%d\n", looplabels[len(looplabels) - 1])
			fmt.Printf(".endloop%d:\n", looplabels[len(looplabels) - 1])
			looplabels = looplabels[:len(looplabels) - 1]

		default:
			panic("invalid word")
		}
	}

	print("		;; EXIT")
	print("		mov rdi, 0")
	print("		mov rax, 60")
	print("		syscall")

	print(".dump:");
	print("		push    rbp");
	print("		mov     rbp, rsp");
	print("		sub     rsp, 48");
	print("		mov     DWORD  [rbp-36], edi");
	print("		mov     QWORD  [rbp-32], 0");
	print("		mov     QWORD  [rbp-24], 0");
	print("		mov     DWORD  [rbp-16], 0");
	print("		mov     BYTE  [rbp-13], 10");
	print("		mov     DWORD  [rbp-4], 18");
	print("		mov     DWORD  [rbp-8], 0");
	print("		cmp     DWORD  [rbp-36], 0");
	print("		jns     .L3");
	print("		neg     DWORD  [rbp-36]");
	print("		mov     DWORD  [rbp-8], 1");
	print("		.L3:");
	print("		mov     edx, DWORD  [rbp-36]");
	print("		movsx   rax, edx");
	print("		imul    rax, rax, 1717986919");
	print("		shr     rax, 32");
	print("		mov     ecx, eax");
	print("		sar     ecx, 2");
	print("		mov     eax, edx");
	print("		sar     eax, 31");
	print("		sub     ecx, eax");
	print("		mov     eax, ecx");
	print("		sal     eax, 2");
	print("		add     eax, ecx");
	print("		add     eax, eax");
	print("		sub     edx, eax");
	print("		mov     DWORD  [rbp-12], edx");
	print("		mov     eax, DWORD  [rbp-12]");
	print("		add     eax, 48");
	print("		mov     edx, eax");
	print("		mov     eax, DWORD  [rbp-4]");
	print("		cdqe");
	print("		mov     BYTE  [rbp-32+rax], dl");
	print("		mov     eax, DWORD  [rbp-12]");
	print("		sub     DWORD  [rbp-36], eax");
	print("		mov     eax, DWORD  [rbp-36]");
	print("		movsx   rdx, eax");
	print("		imul    rdx, rdx, 1717986919");
	print("		shr     rdx, 32");
	print("		mov     ecx, edx");
	print("		sar     ecx, 2");
	print("		cdq");
	print("		mov     eax, ecx");
	print("		sub     eax, edx");
	print("		mov     DWORD  [rbp-36], eax");
	print("		sub     DWORD  [rbp-4], 1");
	print("		cmp     DWORD  [rbp-36], 0");
	print("		jne     .L3");
	print("		cmp     DWORD  [rbp-8], 0");
	print("		je      .L4");
	print("		mov     eax, DWORD  [rbp-4]");
	print("		cdqe");
	print("		mov     BYTE  [rbp-32+rax], 45");
	print("		sub     DWORD  [rbp-4], 1");
	print("		.L4:");
	print("		mov     eax, 20");
	print("		sub     eax, DWORD  [rbp-4]");
	print("		cdqe");
	print("		mov     edx, DWORD  [rbp-4]");
	print("		movsx   rdx, edx");
	print("		lea     rcx, [rbp-32]");
	print("		add     rcx, rdx");
	print("		mov     rdx, rax");
	print("		mov     rsi, rcx");
	print("		mov     edi, 1");
	print("		mov 	rax, 1");
	print("		syscall");
	print("		nop");
	print("		leave");
	print("		ret")
}

func main() {
	argv := os.Args;
	argc := len(argv);

	if (argc != 2) {
		panic("Invaid usage");
	}

	raw := readfile(argv[argc - 1])
	compile(raw);
}
