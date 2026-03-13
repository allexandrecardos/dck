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

## Comandos principais

```bash
# listar containers com seleção interativa
dck ps

# subir serviços (docker compose up -d)
dck up

# subir em foreground
dck up -f

# mostrar o que vai acontecer (docker compose --dry-run up -d)
dck up --dry

# parar e remover serviços
dck down

# executar shell em um container (com UI)
dck exec

# logs interativos
dck logs

# limpeza guiada
dck clean
```

## Licença

MPL-2.0

## Autor

Allexandre Cardoso (@allexandrecardos)
