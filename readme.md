# HLP: a OpenAI GPT CLI Tool

This command-line tool allows users to interact with OpenAI's GPT-3 API to generate responses to questions and save chat logs to a file. The tool has three subcommands: "ask", "auth", and "chat".

## Installation

To install the tool, run the following command:

```bash
go install github.com/yiblet/hlp@latest
```

## Subcommands

### Ask

The "ask" subcommand takes a question as input and uses the GPT-3 API to generate a response. Users can specify parameters such as the maximum number of tokens and the temperature for generating the response. The response will be output to the console.

```bash
hlp ask "How do I recursively alter all files to the standard chmod permissions in a directory?"
```

### Auth

The "auth" subcommand allows users to store their OpenAI API key for use with the tool. If the API key is not passed in as an environment variable or command line argument, the user will be prompted to enter it.

```bash
hlp auth
```

### Chat

The "chat" subcommand takes a chat log file as input and generates responses from the GPT-3 API for each message in the log. The chat log file should be formatted with alternating roles and content, separated by a line containing "---". The tool will append the generated responses to the chat log file.

```bash
hlp chat chat.log
```

The chat log file should have the following format:

```
--- role1
Content for role1
(additional lines of content for Role1 if necessary)
--- role2
Content for role2
(additional lines of content for Role2 if necessary)
```

The file should have alternating roles and content, separated by a line containing `---`. Each role and its corresponding content must be separated by a newline. Valid roles are "system", "assistant", and "user". The "system" role can only appear as the first role in the chat log.

When you pass "-" into the input file, the tool will read from `stdin` instead. When you pass "-" into the output file, the tool will output the results to `stdout` instead of writing to a file. This can be useful for piping the output of one command to the input of another.

## Configuration

The tool requires an OpenAI API key to be configured for use with the subcommands. The API key can be passed in as an environment variable or command line argument. If the API key is not configured, the "auth" subcommand can be used to store the API key.

## Dependencies

The tool is written in Go and imports the "go-gpt3" and "go-arg" packages.

## Contributing

If you would like to contribute to the project, please fork the repository and submit a pull request. Contributions are welcome and appreciated.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
