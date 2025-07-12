data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}


# ここを追加
variable "ssh_key" {
  description = "EC2 で使う SSH Key"
  type        = string
}

# key 名は任意の名前で良い。GitHub の key と同じように作業しているPCの名前がおすすめ
resource "aws_key_pair" "honahuku_thinkpad_20250621" {
  key_name   = "honahuku-thinkpad-20250621"
  public_key = var.ssh_key
}

resource "aws_instance" "kdg-aws-20250621" {
  ami = data.aws_ami.ubuntu.id
  # AWS の無力枠を使いたいため t3.micro を使う
  instance_type = "t3.micro"

  tags = {
    Name     = "kdg-aws-20250621",
    UserDate = "true"
  }

  vpc_security_group_ids = [aws_security_group.ssh_enable.id]

  key_name = aws_key_pair.honahuku_thinkpad_20250621.id # ここを追加
}




