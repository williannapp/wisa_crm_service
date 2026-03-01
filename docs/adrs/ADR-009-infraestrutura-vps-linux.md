# ADR-009 — Infraestrutura VPS Linux com Hardening

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (infraestrutura de servidor)

---

## Contexto

O `wisa-crm-service` opera como Identity Provider centralizado — toda a cadeia de autenticação da plataforma SaaS depende da disponibilidade e segurança desta infraestrutura. O sistema roda em uma **VPS Linux**, que é um servidor virtual privado compartilhando hardware físico com outros clientes do provedor de cloud.

O ambiente VPS tem características específicas que impactam o design de segurança:

- O servidor é acessível via internet pública
- O hipervisor e o hardware são compartilhados (diferente de bare-metal dedicado)
- O sistema operacional e seus serviços são de responsabilidade do operador (não do provedor)
- O IP da VPS é fixo e conhecido — alvo potencial para ataques direcionados
- A chave privada RSA para assinar JWT está armazenada neste servidor — ativo crítico

Um servidor comprometido significa comprometimento total da plataforma: a chave privada RSA pode ser exfiltrada, permitindo forjar tokens para qualquer tenant.

---

## Decisão

**A VPS Linux (Ubuntu LTS 24.04) será configurada seguindo um processo rigoroso de hardening do sistema operacional e dos serviços.** O objetivo é minimizar a superfície de ataque ao mínimo necessário para operar os serviços (`wisa-crm-service` + NGINX + PostgreSQL).

---

## Justificativa

### 1. Princípio do menor privilégio em todos os níveis

**Usuários do sistema:**
- Cada serviço roda com seu próprio usuário dedicado (sem privilégios de root):
  - `wisa-crm` para o backend Go
  - `www-data` para NGINX
  - `postgres` para PostgreSQL
- Acesso SSH permitido apenas para usuário não-root com sudo restrito
- Root login via SSH desabilitado

**Portas abertas (regra: se não é necessário, não está aberto):**

| Porta | Serviço | Visibilidade |
|-------|---------|-------------|
| 80/tcp | NGINX (redirect para HTTPS) | Internet |
| 443/tcp | NGINX (HTTPS) | Internet |
| 22/tcp | SSH | Restrito (whitelist de IPs) |
| 5432/tcp | PostgreSQL | Loopback apenas (127.0.0.1) |
| 8080/tcp | Backend Go | Loopback apenas (127.0.0.1) |

**O PostgreSQL e o backend Go nunca são acessíveis diretamente da internet.**

### 2. Configuração de Firewall (UFW / iptables)

```bash
# UFW — política padrão: deny all
ufw default deny incoming
ufw default allow outgoing

# Permitir SSH apenas de IPs conhecidos (exemplos)
ufw allow from 203.0.113.0/24 to any port 22

# HTTP e HTTPS públicos
ufw allow 80/tcp
ufw allow 443/tcp

# Habilitar firewall
ufw enable

# Verificar regras
ufw status verbose
```

**Proteção contra ataques comuns com iptables:**

```bash
# Rate limiting de novas conexões SSH (previne brute force de SSH)
iptables -A INPUT -p tcp --dport 22 -m state --state NEW -m recent --set
iptables -A INPUT -p tcp --dport 22 -m state --state NEW -m recent --update --seconds 60 --hitcount 4 -j DROP

# Proteção contra SYN flood
iptables -A INPUT -p tcp --syn -m limit --limit 1/s --limit-burst 3 -j ACCEPT
iptables -A INPUT -p tcp --syn -j DROP

# Bloquear pacotes inválidos
iptables -A INPUT -m state --state INVALID -j DROP
```

### 3. Hardening de SSH

```
# /etc/ssh/sshd_config

Protocol 2
Port 22                          # Considerar porta não-padrão como ofuscação adicional
PermitRootLogin no               # Root nunca via SSH
PasswordAuthentication no        # Apenas chave pública
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
PermitEmptyPasswords no
X11Forwarding no
AllowTcpForwarding no
MaxAuthTries 3
LoginGraceTime 20
ClientAliveInterval 300
ClientAliveCountMax 2
AllowUsers deploy-user           # Whitelist de usuários SSH
```

**Autenticação SSH apenas com chave RSA 4096-bit ou Ed25519.** Passwords desabilitados completamente.

### 4. Systemd para gerenciamento dos serviços

Cada serviço roda como uma unidade systemd com restrições de segurança adicionais:

```ini
# /etc/systemd/system/wisa-crm.service

[Unit]
Description=WISA CRM Authentication Service
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=wisa-crm
Group=wisa-crm
WorkingDirectory=/opt/wisa-crm
ExecStart=/opt/wisa-crm/wisa-crm-service
Restart=on-failure
RestartSec=5s

# Hardening de systemd
NoNewPrivileges=true              # Processo não pode ganhar novos privilégios
PrivateTmp=true                   # /tmp privado e isolado
ProtectSystem=strict              # Filesystem somente leitura (exceto caminhos explícitos)
ProtectHome=true                  # Sem acesso a /home
ReadWritePaths=/var/log/wisa-crm  # Somente este path para escrita
CapabilityBoundingSet=            # Nenhuma capability de kernel
AmbientCapabilities=              # Nenhuma capability ambiente
SecureBits=noroot                 # Previne escalação para root
MemoryMax=512M                    # Limite de memória

# Variáveis de ambiente via arquivo protegido
EnvironmentFile=/etc/wisa-crm/env

[Install]
WantedBy=multi-user.target
```

As restrições do systemd garantem que mesmo se o processo Go for comprometido via RCE, o atacante terá acesso mínimo ao sistema.

### 5. Gestão segura da chave privada RSA

A chave privada é o ativo mais crítico do servidor. Sua proteção:

```bash
# Diretório com permissão restrita
mkdir -p /etc/wisa-crm/keys
chown root:wisa-crm /etc/wisa-crm/keys
chmod 750 /etc/wisa-crm/keys

# Chave privada — somente o usuário do serviço lê
chmod 400 /etc/wisa-crm/keys/private.pem
chown wisa-crm:wisa-crm /etc/wisa-crm/keys/private.pem
```

**A chave privada:**
- Nunca é armazenada no repositório git (`.gitignore` + `git-secrets`)
- Nunca aparece em logs
- É gerada diretamente no servidor, não transferida via rede
- É referenciada via path absoluto na configuração do serviço
- Tem backup criptografado em local seguro fora da VPS

### 6. Atualizações automáticas de segurança

```bash
# Instalar unattended-upgrades
apt install unattended-upgrades

# /etc/apt/apt.conf.d/50unattended-upgrades
Unattended-Upgrade::Allowed-Origins {
    "${distro_id}:${distro_codename}-security";
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Automatic-Reboot "false";  # Reboot manual para controle
Unattended-Upgrade::Mail "admin@wisa-crm.com";
```

Patches de segurança do OS são aplicados automaticamente. Reboots são manuais e planejados (exceto em emergências críticas).

### 7. Fail2ban para proteção contra força bruta

```ini
# /etc/fail2ban/jail.local

[sshd]
enabled = true
maxretry = 3
bantime = 3600   # 1 hora

[nginx-http-auth]
enabled = true

[nginx-botsearch]
enabled = true
maxretry = 2
bantime = 86400  # 1 dia

# Jail customizado para endpoint de login
[wisa-crm-login]
enabled = true
filter = wisa-crm-login
logpath = /var/log/wisa-crm/access.log
maxretry = 10
findtime = 300   # 5 minutos
bantime = 3600   # 1 hora
```

```ini
# /etc/fail2ban/filter.d/wisa-crm-login.conf
[Definition]
failregex = ^.*"POST /api/v1/auth/login.*" 401
```

### 8. Monitoramento e observabilidade

**Stack mínima de monitoramento:**

- **Prometheus + Node Exporter:** métricas de sistema (CPU, memória, disco, rede)
- **Prometheus + custom metrics do backend Go:** goroutines ativas, requisições/segundo, latência de login, tokens emitidos
- **Alertmanager:** alertas por email/Slack para:
  - CPU > 80% por > 5 minutos
  - Memória > 85%
  - Disco > 80%
  - Certificado SSL expirando em < 30 dias
  - Serviço `wisa-crm` restartando repetidamente
  - Taxa de erro de login > threshold (possível ataque)

**Logs centralizados:**
- Logs do NGINX: `/var/log/nginx/wisa-crm-access.log` e `error.log`
- Logs do backend Go: estruturado em JSON para facilitar parsing
- Logs do PostgreSQL: configurado conforme ADR-003
- Rotação com `logrotate`: manter 30 dias, comprimir após 7 dias

### 9. Backup e recuperação

```bash
# Backup diário do banco de dados
# /etc/cron.d/wisa-crm-backup

0 2 * * * postgres pg_dump -Fc wisa_crm_db > /backups/wisa_crm_$(date +\%Y\%m\%d).dump
0 3 * * * find /backups -name "*.dump" -mtime +30 -delete
```

**Backup offsite:** copiar backups para bucket S3 ou servidor externo (rsync + SSH):
```bash
0 4 * * * rsync -avz /backups/ backup-server:/wisa-crm-backups/
```

**Testar restore mensalmente** com documentação do processo e tempo esperado de Recovery Time Objective (RTO).

---

## Consequências

### Positivas
- Superfície de ataque mínima: apenas portas 80, 443 e 22 (restrita) expostas
- Isolamento de processos via systemd impede escalação de privilégios mesmo em caso de RCE
- Hardening de SSH elimina ataques de força bruta por senha
- Atualizações automáticas de segurança reduzem janela de exposição a CVEs
- Monitoramento detecta anomalias antes de impacto ao usuário

### Negativas
- Configuração rigorosa aumenta complexidade do processo de deploy inicial
- Restrições do systemd podem precisar de ajuste fino se o serviço precisar de capabilities específicas
- Alta dependência de uma única VPS — sem alta disponibilidade nativa
- Operação de um servidor Linux exige conhecimento de sysadmin da equipe

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Chave SSH de acesso comprometida | Baixa | Crítico | Alta |
| CVE crítico no kernel Linux antes do patch | Baixa | Alto | Alta |
| VPS provider comprometido (side-channel attack) | Muito Baixa | Crítico | Média |
| Disco cheio causando indisponibilidade | Média | Alto | Alta |
| Falha de hardware na VPS sem SLA de recuperação | Baixa | Crítico | Alta |
| Acesso físico ao hipervisor pelo provider | Muito Baixa | Crítico | Média |

---

## Mitigações

### Chave SSH comprometida
- Usar chaves Ed25519 (mais seguras e curtas) ou RSA 4096
- Proteger a chave privada SSH com passphrase
- Rotacionar chaves SSH a cada 6 meses ou imediatamente ao desligamento de um membro da equipe
- Manter lista de chaves autorizadas (`authorized_keys`) versionada e auditada

### CVE no kernel
- Atualizações automáticas de segurança (configurado acima)
- Inscrever-se no Ubuntu Security Notices para alertas de CVEs críticos
- Política de reboot de emergência: definir janela de manutenção mensal para aplicar kernel updates que requerem reboot

### Disco cheio
- Monitorar `/` com alerta em 80% e crítico em 90%
- Configurar logrotate para evitar crescimento descontrolado de logs
- Monitorar crescimento do banco de dados com `pg_database_size()`

### Falha de hardware (HA básica)
- **Snapshot da VPS:** configurar snapshots diários automáticos no painel do provedor de cloud
- **Backup offsite:** banco de dados em servidor externo (conforme seção de backup)
- **Runbook de recuperação:** documentar processo de provisionar nova VPS e restaurar serviço em caso de falha total
- RTO (Recovery Time Objective) alvo: < 2 horas

---

## Checklist de Hardening (para implantação)

```
[ ] Ubuntu LTS atualizado (apt update && apt upgrade)
[ ] Usuário não-root criado para administração
[ ] SSH configurado (sem root, sem senha, chave pública)
[ ] UFW configurado (deny all, allow 22/whitelist, 80, 443)
[ ] Fail2ban instalado e configurado
[ ] unattended-upgrades configurado
[ ] Usuários de serviço criados (wisa-crm, sem home, sem shell)
[ ] Diretório de chaves criado com permissões restritas
[ ] Chave RSA gerada diretamente no servidor
[ ] systemd service configurado com hardening
[ ] NGINX instalado e configurado (ADR-007)
[ ] PostgreSQL instalado e configurado (ADR-003)
[ ] PgBouncer instalado
[ ] Monitoramento instalado (Node Exporter + Prometheus)
[ ] Alertas configurados (certificado, disco, CPU, memória)
[ ] Backup configurado (cron + offsite)
[ ] Restore testado
[ ] git-secrets ou gitleaks configurado na CI
[ ] Varredura de portas abertas verificada (nmap da internet)
```

---

## Alternativas Consideradas

### Kubernetes em VPS (k3s)
- **Prós:** Orquestração de containers, restart automático, rolling updates
- **Contras:** Overhead de complexidade enorme para um único serviço; k3s ainda tem superfície de ataque significativa; requer conhecimento especializado; custo de overhead de CPU/memória

### Docker Compose
- **Prós:** Isolamento de containers, imagens reproduzíveis, rollback simples
- **Contras:** Adiciona camada de abstração sem benefício claro para um serviço único; rede entre containers adiciona complexidade; não elimina necessidade de hardening do host; runtime Docker adiciona superfície de ataque

### Bare-metal dedicado
- **Prós:** Isolamento de hardware total, sem risco de side-channel de hipervisor
- **Contras:** Custo significativamente maior; provisionamento mais lento; sem snapshots automáticos; justificado apenas em ambientes de compliance estrito (PCI-DSS, HIPAA)

**Para o estágio atual, VPS Linux com hardening sistemático oferece o equilíbrio correto entre segurança, custo e operabilidade.**

---

## Referências

- [Ubuntu CIS Benchmark](https://www.cisecurity.org/benchmark/ubuntu_linux)
- [SSH Hardening Guide](https://infosec.mozilla.org/guidelines/openssh)
- [Systemd Service Hardening](https://www.freedesktop.org/software/systemd/man/latest/systemd.exec.html)
- [Fail2ban Documentation](https://www.fail2ban.org/wiki/index.php/MANUAL_0_8)
- [UFW Documentation](https://help.ubuntu.com/community/UFW)
- [OWASP Infrastructure Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Infrastructure_as_Code_Security_Cheat_Sheet.html)
