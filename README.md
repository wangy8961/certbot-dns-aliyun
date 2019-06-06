# certbot-dns-aliyun
用 Let's Encrypt 官方工具 Certbot 申请通配符证书（Wildcard Certificate）时，只能用 DNS-01 的方式来验证域名所有权，需要在域名下添加一条 DNS TXT 记录。如果要用 certbot renew 命令自动续期的话，就需要自动添加或删除 DNS TXT 记录。官方提供的都是国外的 DNS 服务商的插件，而国内的 Aliyun DNS 也提供了 DNS 云解析管理 API，此工具是用 Go 语言调用 API 实现自动添加和删除 DNS TXT 记录，从而实现自动用 certbot renew 命令续期通配符证书的目的！
