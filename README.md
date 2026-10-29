Monkey Language Interpreter (a programming language made only in Go)
---

What is this?

An interpreter for a C-like programming language from scratch with a lexer, a parser and an Abstract Syntax Tree (AST). With Pratt parsing technique and a recursive descent parser. 
A project for myself to learn about how compilers work! To learn What closures are and how and why they work, what are built-in data structures, and What REPL stands for and how to build one

---

## How it works
```
Source Code (string)
       ↓
    Lexer          → breaks input into tokens  (e.g. `let`, `five`, `=`, `5`)
       ↓
    Parser         → builds an Abstract Syntax Tree (AST)
       ↓
   Evaluator       → walks the AST and executes it
       ↓
    Output
```

---

## The Monkey Language

Monkey supports:

- Variable bindings: `let x = 5;`
- Functions: `let add = fn(x, y) { x + y; };`
- Arithmetic & comparisons: `!-/*5; 5 < 10 > 5;`
- A fully working **REPL** (Read-Eval-Print Loop)

---

---
