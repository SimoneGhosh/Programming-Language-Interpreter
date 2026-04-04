# Monkey Language Interpreter

> Building a programming language from scratch in Go — following *Writing an Interpreter in Go* by Thorsten Ball

---

## What is this?

This is a hand-rolled interpreter for **Monkey**, a C-like scripting language, built entirely from scratch in Go — no parser generators, no libraries. Every component is implemented manually: the lexer, the token system, the parser, and the REPL.

This project is about understanding what happens *under the hood* when code runs — turning raw text into something a machine can execute, one character at a time.

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

## Try it yourself
```bash
cd Programming-Language-Interpreter/lexing/src/monkey
go run main.go
```
```
You're in the REPL! Write your code :)
>> let add = fn(x, y) { x + y; };
>> add(5, 10)
15
```

---

## Run the tests
```bash
go test ./lexer

```

```See how it works
go test -v -run TestOperatorPrecedenceParsing ./parser
```

---

## Tech

- **Language:** Go
- **Concepts:** Lexing, parsing, AST construction, tree-walking evaluation
- **Reference:** *Writing an Interpreter in Go* — Thorsten Ball
