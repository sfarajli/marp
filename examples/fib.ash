# Print out given number of elements in the Fibonacci sequence, 46 is the max value
define fib
	var lim push lim
	var a 1 push a
	var b 1 push b
	pull lim while dup 0 > do
		pull a dump
		pull b dup
		pull a
		+ push b
		push a
		1 -
	done
end

10 fib
