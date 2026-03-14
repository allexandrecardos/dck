# DCK Helper CLI (`dck`)

`dck` é uma CLI produtiva para trabalhar com Docker e Docker Compose, com comandos interativos e atalhos para o dia a dia.

## Instalação

### Windows

```powershell
# Instalar (última versão)
irm https://raw.githubusercontent.com/allexandrecardos/dck/main/install/install.ps1 | iex

# Teste
dck version
```

Para instalar uma versão específica:

```powershell
$env:DCK_VERSION="v0.1.0"; irm https://raw.githubusercontent.com/allexandrecardos/dck/main/install/install.ps1 | iex
```

Se o diretório padrão não tiver permissão, o instalador usa `%USERPROFILE%\dck` e adiciona ao PATH do usuário.

### Linux

```bash
# Instalar (última versão)
curl -fsSL https://raw.githubusercontent.com/allexandrecardos/dck/main/install/install.sh | sh

# Teste
dck version
```

Para instalar uma versão específica:

```bash
DCK_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/allexandrecardos/dck/main/install/install.sh | sh
```

Você pode sobrescrever o destino com `DCK_INSTALL_DIR`:

```bash
DCK_INSTALL_DIR=$HOME/.local/bin curl -fsSL https://raw.githubusercontent.com/allexandrecardos/dck/main/install/install.sh | sh
```

## Atualização

```bash
# checar versão mais recente
dck update --check

# atualizar automaticamente
dck update
```

## Desinstalação

```bash
# remove o binário
dck uninstall

# remove binário e dck-config.yml
dck uninstall --purge
```

No Windows, a remoção é agendada após o comando terminar porque o executável está em uso.

## Configuração

O arquivo de configuração é criado no mesmo diretório onde o `dck` está instalado:

```
<install-dir>/dck-config.yml
```

Abra/edite com:

```bash
dck config
```

## Comandos e exemplos

**ps**
Lista containers em tabela (colunas configuráveis no `dck-config.yml`).

```bash
dck ps
dck ps -a
```

**logs**
Mostra logs do container (com seleção interativa se não passar nome).

```bash
dck logs api
dck logs -f -n 200
dck logs
```

**exec**
Entra em um container com shell detectado automaticamente.

```bash
dck exec api
dck exec -s /bin/sh api
dck exec -c "ls -la" api
```

**run**
Cria e inicia um container com defaults seguros. Faz pull automático da imagem se necessário.

```bash
dck run nginx
dck run nginx -p 8080:80
dck run postgres:16 --name db -e POSTGRES_PASSWORD=secret
dck run ubuntu bash
```

**start**
Inicia containers parados (com seleção interativa se não passar args).

```bash
dck start api worker
dck start
dck start -a
```

**stop**
Para containers em execução (com seleção interativa se não passar args).

```bash
dck stop api worker
dck stop
dck stop -a
dck stop -t 5 api
```

**pause**
Pausa containers em execução.

```bash
dck pause api worker
dck pause
dck pause -a
```

**rm**
Remove recursos Docker (containers, imagens, volumes, networks).

```bash
dck rm api
dck rm -i nginx:latest
dck rm --deep api
dck rm -y
dck rm
```

**inspect**
Mostra o JSON do `docker inspect`.

```bash
dck inspect api
dck inspect
```

**clean**
Limpa recursos Docker não utilizados (com confirmação).

```bash
dck clean
```

**up**
Sobe serviços com `docker compose up`.

```bash
dck up
dck up -f
dck up --dry
```

**down**
Derruba serviços com `docker compose down`.

```bash
dck down
```

**init**
Cria `Dockerfile` e `docker-compose.yml` no diretório atual.

```bash
dck init
dck init -b
```

**config**
Abre (ou cria) o `dck-config.yml` no diretório de instalação.

```bash
dck config
```

**update**
Verifica e/ou atualiza para a versão mais recente.

```bash
dck update --check
dck update
```

**version**
Mostra a versão instalada.

```bash
dck version
```

**uninstall**
Remove o binário e opcionalmente o config.

```bash
dck uninstall
dck uninstall --purge
```

## Licença

MPL-2.0

## Autor

Allexandre Cardoso (@allexandrecardos)
