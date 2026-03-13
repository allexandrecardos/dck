# DCK Helper CLI (`dck`)

`dck` é uma CLI produtiva para trabalhar com Docker e Docker Compose, com comandos interativos e atalhos para o dia a dia.

## Instalação

### Windows

```powershell
# Baixar e instalar (releases)
.\install\install.ps1

# Teste
dck version
```

Se o diretório padrão não tiver permissão, o instalador usa `%USERPROFILE%\dck` e adiciona ao PATH do usuário.

### Linux

Instalação com `curl`:

```bash
curl -fsSL https://raw.githubusercontent.com/allexandrecardos/dck/main/install/install.sh | sh
```

Ou rodando o script localmente:

```bash
# Baixar e instalar (releases)
./install/install.sh

# Teste
dck version
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
