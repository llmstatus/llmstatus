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
  key_name   = "${var.project}-${var.environment}-main"
  public_key = var.ssh_public_key
}

resource "aws_security_group" "this" {
  name        = "${var.project}-${var.environment}-main"
  description = "llmstatus main server: HTTP/HTTPS/SSH inbound"

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  dynamic "ingress" {
    for_each = length(var.probe_node_cidrs) > 0 ? [1] : []
    content {
      description = "Postgres from probe nodes"
      from_port   = 15432
      to_port     = 15432
      protocol    = "tcp"
      cidr_blocks = var.probe_node_cidrs
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-main"
  })
}

resource "aws_instance" "this" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.small"
  key_name               = aws_key_pair.this.key_name
  vpc_security_group_ids = [aws_security_group.this.id]

  root_block_device {
    volume_type = "gp3"
    volume_size = 20
    encrypted   = true

    tags = merge(var.tags, {
      Name = "${var.project}-${var.environment}-main-root"
    })
  }

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-main"
    Role = "main"
  })
}

resource "aws_ebs_volume" "data" {
  availability_zone = aws_instance.this.availability_zone
  size              = var.data_volume_size_gb
  type              = "gp3"
  encrypted         = true

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-data"
  })
}

resource "aws_volume_attachment" "data" {
  device_name  = "/dev/sdf"
  volume_id    = aws_ebs_volume.data.id
  instance_id  = aws_instance.this.id
  force_detach = false
}

resource "aws_eip" "this" {
  domain = "vpc"

  tags = merge(var.tags, {
    Name = "${var.project}-${var.environment}-main"
  })
}

resource "aws_eip_association" "this" {
  instance_id   = aws_instance.this.id
  allocation_id = aws_eip.this.id
}
