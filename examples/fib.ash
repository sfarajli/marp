# Print out given number of elements in the Fibonacci sequence, 46 is the max value
define fib
	var lim !lim
	var a 1 !a
	var b 1 !b
	@lim while dup 0 > do
		@a dump
		@b dup
		@a
		+ !b
		!a
		1 -
	done
end

10 fib
