package gocamole

import "fmt"

type Instruction struct {
	Opcode string
	Args   []string
}

func (ins *Instruction) String() string {
	var s string

	s += fmt.Sprintf("%d.%s", len(ins.Opcode), ins.Opcode)
	if ins.Args != nil {
		for _, arg := range ins.Args {
			s += fmt.Sprintf(",%d.%s", len(arg), arg)
		}
	}

	s += ";"

	return s
}

func (ins *Instruction) Bytes() []byte {
	return []byte(ins.String())
}

func selectHandshakeInstruction(protocol string) *Instruction {
	return &Instruction{
		Opcode: "select",
		Args:   []string{protocol},
	}
}

func sizeHandshakeInstruction(width string, height string, dpi string) *Instruction {
	return &Instruction{
		Opcode: "size",
		Args:   []string{width, height, dpi},
	}
}

func audioHandshakeInstruction(mimetypes []string) *Instruction {
	return &Instruction{
		Opcode: "audio",
		Args:   mimetypes,
	}
}

func videoHandshakeInstruction(mimetypes []string) *Instruction {
	return &Instruction{
		Opcode: "video",
		Args:   mimetypes,
	}
}

func imageHandshakeInstruction(mimetypes []string) *Instruction {
	return &Instruction{
		Opcode: "image",
		Args:   mimetypes,
	}
}

func connectHandshakeInstruction(argIns *Instruction, config *Config) *Instruction {
	args := make([]string, 0)
	for i, arg := range argIns.Args {
		// The first argument is the protocol version
		if i == 0 {
			args = append(args, arg)
			continue
		}

		switch arg {
		case "hostname":
			args = append(args, config.Hostname)
		case "port":
			args = append(args, config.Port)
		case "username":
			args = append(args, config.Username)
		case "password":
			args = append(args, config.Password)
		case "security":
			args = append(args, config.Security)
		case "ignore-cert":
			args = append(args, config.IgnoreCert)
		default:
			args = append(args, "")
		}
	}

	return &Instruction{
		Opcode: "connect",
		Args:   args,
	}
}
