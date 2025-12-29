# Notas para Assistentes de IA (Claude Code & Gemini CLI)

**⚠️ IMPORTANTE: Este é o arquivo FONTE DA VERDADE para todos os assistentes de IA.**
**Qualquer alteração nas instruções deve ser feita SOMENTE neste arquivo.**

---

## Projeto WABAGO - Waybar Configuration TUI Editor

### Requisitos Implementados

#### Interface e Experiência do Usuário (UI/UX)
- [x] **Placeholders Inteligentes no Editor de Módulos**: Adição de placeholders nos campos de atributos que não possuem valor definido.
    - Exibe valores possíveis para enums/opções específicas (ex: "top, bottom, left, right").
    - Exibe "true, false" para campos booleanos.
    - Estilização diferenciada com cor esmaecida para distinguir de valores reais.
    - Truncamento inteligente para placeholders longos (máx. 40 caracteres).
- [x] **Estilização de Formulários**: Criação de `PlaceholderStyle` para consistência visual.

#### Sistema de Backup e Segurança
- [x] **Backup Temporário de Inicialização**: Criação automática de um backup do estado atual (`config` e `style`) ao iniciar a aplicação.
- [x] **Timestamps em Backups de Inicialização**: Inclusão de carimbo de tempo (formato `YYYYMMDD_HHMMSS`) no nome dos backups automáticos para evitar colisões.
- [x] **Limpeza Automática (Cleanup)**: Remoção dos backups temporários de inicialização ao encerrar a aplicação, mantendo o diretório de backups limpo.

#### Estabilidade e Manutenção
- [x] **Resolução de Conflitos de Compilação**: Limpeza de declarações duplicadas em `internal/ui` para garantir build estável.
- [x] **Integração no Ciclo de Vida**: Uso de `defer` no `main.go` para garantir a execução do Cleanup.

### Referências e Fonte da Verdade - Waybar

As informações contidas nos links abaixo devem ser consideradas como a **fonte da verdade** para quaisquer implementações, validações de atributos e configurações de módulos. Consulte-os para obter detalhes sobre propriedades, tipos de dados permitidos e comportamentos específicos.

#### Configuração Geral e Estilização
- [Configuration](https://github.com/Alexays/Waybar/wiki/Configuration): Visão geral da estrutura do arquivo de configuração, posições, camadas e opções globais.
- [Styling](https://github.com/Alexays/Waybar/wiki/Styling): Guia completo sobre seletores CSS, classes e IDs para estilizar a barra e seus módulos.
- [Writing Modules](https://github.com/Alexays/Waybar/wiki/Writing-Modules): Guia para desenvolvedores sobre como criar novos módulos personalizados para o Waybar.

#### Módulos de Hardware e Sistema
- [Backlight](https://github.com/Alexays/Waybar/wiki/Module:-Backlight): Exibe e controla o nível de brilho da tela.
- [Backlight Slider](https://github.com/Alexays/Waybar/wiki/Module:-Backlight-Slider): Slider interativo para ajuste de brilho.
- [Battery](https://github.com/Alexays/Waybar/wiki/Module:-Battery): Monitoramento de bateria, estados de carga e formatação de tempo.
- [Bluetooth](https://github.com/Alexays/Waybar/wiki/Module:-Bluetooth): Status do adaptador Bluetooth, dispositivos conectados e controle.
- [CPU](https://github.com/Alexays/Waybar/wiki/Module:-CPU): Uso da CPU, carga e temperatura (se suportado).
- [Disk](https://github.com/Alexays/Waybar/wiki/Module:-Disk): Uso de espaço em disco e estatísticas de montagem.
- [Load](https://github.com/Alexays/Waybar/wiki/Module:-Load): Monitoramento da carga média do sistema (load average).
- [Memory](https://github.com/Alexays/Waybar/wiki/Module:-Memory): Uso de memória RAM e SWAP.
- [Network](https://github.com/Alexays/Waybar/wiki/Module:-Network): Status de interfaces de rede (WiFi, Ethernet), IP e velocidade.
- [PowerProfilesDaemon](https://github.com/Alexays/Waybar/wiki/Module:-PowerProfilesDaemon): Controle de perfis de energia (performance, balanced, power-saver).
- [Temperature](https://github.com/Alexays/Waybar/wiki/Module:-Temperature): Monitoramento de sensores de temperatura do sistema.
- [UPower](https://github.com/Alexays/Waybar/wiki/Module:-UPower): Abstração para dispositivos de energia suportados pelo UPower.

#### Áudio e Mídia
- [Cava](https://github.com/Alexays/Waybar/wiki/Module:-Cava): Visualizador de áudio (barras) integrado.
- [Cava (Raw/GLSL)](https://github.com/Alexays/Waybar/wiki/Module:-Cava:-Raw) / [GLSL](https://github.com/Alexays/Waybar/wiki/Module:-Cava:-GLSL): Configurações avançadas do Cava.
- [JACK](https://github.com/Alexays/Waybar/wiki/Module:-JACK): Status e carga do servidor de áudio JACK.
- [MPD](https://github.com/Alexays/Waybar/wiki/Module:-MPD): Cliente para o Music Player Daemon (status, música atual).
- [MPRIS](https://github.com/Alexays/Waybar/wiki/Module:-MPRIS): Controle genérico de players de mídia (Spotify, VLC, etc.).
- [PulseAudio](https://github.com/Alexays/Waybar/wiki/Module:-PulseAudio): Controle de volume, entradas/saídas e status do PulseAudio.
- [PulseAudio Slider](https://github.com/Alexays/Waybar/wiki/Module:-PulseAudio-Slider): Slider para volume do PulseAudio.
- [Sndio](https://github.com/Alexays/Waybar/wiki/Module:-Sndio): Controle de volume para o servidor de áudio Sndio.
- [WirePlumber](https://github.com/Alexays/Waybar/wiki/Module:-WirePlumber): Controle de áudio para sessões PipeWire via WirePlumber.

#### Gerenciadores de Janelas (Compositors)
- [Hyprland](https://github.com/Alexays/Waybar/wiki/Module:-Hyprland): Módulos específicos para Hyprland (workspaces, janelas, submaps).
- [Sway](https://github.com/Alexays/Waybar/wiki/Module:-Sway): Módulos específicos para Sway (workspaces, mode, window).
- [River](https://github.com/Alexays/Waybar/wiki/Module:-River): Tags e layout para o compositor River.
- [Dwl](https://github.com/Alexays/Waybar/wiki/Module:-Dwl): Tags para o compositor dwl.
- [Niri](https://github.com/Alexays/Waybar/wiki/Module:-Niri): Workspaces e janelas para o compositor Niri.
- [Workspaces](https://github.com/Alexays/Waybar/wiki/Module:-Workspaces): Módulo genérico de workspaces (compatível com vários compositores).
- [Taskbar](https://github.com/Alexays/Waybar/wiki/Module:-Taskbar): Lista de janelas abertas (estilo barra de tarefas wlr).

#### Interface e Utilitários
- [Clock](https://github.com/Alexays/Waybar/wiki/Module:-Clock): Data, hora e calendário.
- [Custom](https://github.com/Alexays/Waybar/wiki/Module:-Custom): Execução de scripts personalizados e exibição de saída.
- [Custom Examples](https://github.com/Alexays/Waybar/wiki/Module:-Custom:-Examples) / [Third-party](https://github.com/Alexays/Waybar/wiki/Module:-Custom:-Third-party): Exemplos práticos de scripts.
- [Gamemode](https://github.com/Alexays/Waybar/wiki/Module:-Gamemode): Indicador de status do Feral GameMode.
- [Group](https://github.com/Alexays/Waybar/wiki/Module:-Group): Agrupamento de vários módulos em um contêiner (com gaveta/drawer opcional).
- [Idle Inhibitor](https://github.com/Alexays/Waybar/wiki/Module:-Idle-Inhibitor): Bloqueio de suspensão/protetor de tela.
- [Image](https://github.com/Alexays/Waybar/wiki/Module:-Image): Exibição de imagem estática ou dinâmica.
- [Keyboard State](https://github.com/Alexays/Waybar/wiki/Module:-Keyboard-State): Status de teclas modificadoras (Caps, Num, Scroll Lock).
- [Language](https://github.com/Alexays/Waybar/wiki/Module:-Language): Indicador de layout de teclado atual.
- [Privacy](https://github.com/Alexays/Waybar/wiki/Module:-Privacy): Indicadores de uso de hardware sensível (câmera, microfone).
- [Systemd Failed Units](https://github.com/Alexays/Waybar/wiki/Module:-Systemd-failed-units): Contagem de serviços systemd em estado de falha.
- [User](https://github.com/Alexays/Waybar/wiki/Module:-User): Informações do usuário logado e tempo de atividade.
- [CFFI](https://github.com/Alexays/Waybar/wiki/Module:-CFFI): Interface para chamadas de função C (avançado).

---

*Este arquivo deve ser atualizado sempre que houver novas informações, funcionalidades ou requisitos.*
