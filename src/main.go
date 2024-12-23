package main

import (
	"fmt"
	"strconv"
	"regexp"
	"os"
) 

func readfile(path string) []string {

	var raw []string;
	reg := regexp.MustCompile(`\s`)

	dat, err := os.ReadFile(path);
	if (err != nil) {
		panic("failed to read file");
	}

	/* FIXME: Find a better solution */
	tmp := reg.Split(string(dat), -1);

	for i := range tmp{
		if tmp[i] != "" {
			raw = append(raw, tmp[i]);
		} 
	};

	return raw;
}

func interpret(raw[]string) {
	var err error;
	var number int;
	var stack[]int;
	var skip_to_end bool;

	for i := 0; i < len(raw); i++ {
		length := len(stack);
		/* Conditional looping */
		if (skip_to_end && raw[i] != "end") {
			continue;
		}
		number, err = strconv.Atoi(raw[i]);
		if err == nil {
			stack = append(stack, number);
			continue; 
		}

		switch raw[i] {
		case "+":
			if length < 2 {
				panic("failed to plus: stack underflow");
			}
			stack[length - 2] = stack[length - 1] + stack[length - 2]
			stack = stack[:length - 1];
		case "-":
			if length < 2 {
				panic("failed to minus: stack underflow");
			}
			stack[length - 2] = stack[length - 1] - stack[length - 2]
			stack = stack[:length - 1];
		case "=":
			if length < 2 {
				panic("failed to check equality: stack underflow");
			}
			if (stack[length - 1] == stack[length - 2]) {
				stack = append(stack, -1);
			} else {
				stack = append(stack, 0);
			}

		case ".":
			if length < 1 {
				panic("failed to dump: stack underflow");
			}
			fmt.Println(stack[length - 1]);
			stack = stack[:length - 1];
		case "if":
			if length < 1 {
				panic("failed to if: stack underflow");
			}
			if (stack[length - 1] == 0) {
				skip_to_end = true;
			}
			stack = stack[:length - 1];
		case "end":
			skip_to_end = false;

		default:
			panic("invalid word");
		}
	} 
}

func compile(raw []string) {
	var err error;
	var number int;

	fmt.Println("format ELF64 executable")
	fmt.Println("")
	fmt.Println("_start:")

	for i := 0; i < len(raw); i++ {
		number, err = strconv.Atoi(raw[i]);
		if err == nil {
			fmt.Printf("	push %v\n", number)
			continue; 
		}

		switch raw[i] {
		case "+":
			fmt.Println("	pop r10")
			fmt.Println("	pop rax")
			fmt.Println("	add rax, r10")
			fmt.Println("	push rax")
		case "-":
			fmt.Println("	pop r10")
			fmt.Println("	pop rax")
			fmt.Println("	sub rax, r10")
			fmt.Println("	push rax")
		case ".":
			fmt.Println("	pop rdi")
			fmt.Println("	call _dump")

		default:
			panic("invalid word");
		}
	} 

	fmt.Println("")
	fmt.Println("_end:")
	fmt.Println("	mov rax, 60")
	fmt.Println("	mov rdi, 0")
	fmt.Println("	syscall")

	fmt.Println("")
	fmt.Println("_dump:")
	fmt.Println("	mov rax, rdi")
	fmt.Println("	mov r10, 0")
	fmt.Println("")
	fmt.Println("	dec rsp")
	fmt.Println("	mov byte [rsp], 10")
	fmt.Println("	inc r10")
	fmt.Println("")
	fmt.Println("_prepend_digit:")
	fmt.Println("	mov rdx, 0")
	fmt.Println("	mov rbx, 10")
	fmt.Println("	div rbx")
	fmt.Println("")
	fmt.Println("	add rdx, 48")
	fmt.Println("	dec rsp")
	fmt.Println("	mov [rsp], dl")
	fmt.Println("	inc r10")
	fmt.Println("")
	fmt.Println("	cmp rax, 0")
	fmt.Println("	jne _prepend_digit")
	fmt.Println("")
	fmt.Println("_print_digit:")
	fmt.Println("	mov rax, 1")
	fmt.Println("	mov rdi, 1")
	fmt.Println("	mov rsi, rsp")
	fmt.Println("	mov rdx, r10")
	fmt.Println("	syscall")
	fmt.Println("")
	fmt.Println("	add rsp, r10")
	fmt.Println("	ret")
}

func main() {
	var file string;
	argv := os.Args;
	argc := len(argv);
	compile_flg := false;

	switch argc {
	case 3:
		if argv[1] != "-S" {
			panic("Invalid usage");
		}
		compile_flg = true;
		fallthrough;
	case 2:
		file = argv[argc - 1];
	default:
		panic("Invaid usage");
	}

	output := readfile(file);
	if (compile_flg) {
		compile(output);
	} else {
		interpret(output);
	}
}
