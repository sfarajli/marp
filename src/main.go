package main

import (
	"fmt"
	"regexp"
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

	/* FIXME: Find a better solution */
	tmp := reg.Split(string(dat), -1);

	for i := range tmp{
		if tmp[i] != "" {
			raw = append(raw, tmp[i]);
		}
	};

	return raw;
}

func compile(raw[]string) {
	print("section .text")
	print("global _start")
	print("_start:")
	print("mov rdi, 0")
	print("mov rax, 60")
	print("syscall")
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
