terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_key_pair" "this" {
  key_name   = "${var.project}-${var.environment}-probe-${var.node_name}"
  public_key = var.ssh_public_key
}

resource "aws_security_group" "this" {
  name        = "${var.project}-${var.environment}-probe-${var.node_name}"
  description = "llmstatus probe ${var.node_name}: SSH inbound, all outbound"

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-probe-${var.node_name}"
  })
}

resource "aws_instance" "this" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.micro"
  key_name               = aws_key_pair.this.key_name
  vpc_security_group_ids = [aws_security_group.this.id]

  root_block_device {
    volume_type = "gp3"
    volume_size = 20
    encrypted   = true

    tags = merge(var.tags, {
      Name = "${var.project}-${var.environment}-probe-${var.node_name}-root"
    })
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-probe-${var.node_name}"
    Role = "probe"
    Node = var.node_name
  })
}

resource "aws_eip" "this" {
  domain = "vpc"

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-probe-${var.node_name}"
  })
}

resource "aws_eip_association" "this" {
  instance_id   = aws_instance.this.id
  allocation_id = aws_eip.this.id
}
