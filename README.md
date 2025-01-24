# **Process File API - Documentação**

## 📖 **Visão Geral**

A **Process File API** é uma aplicação projetada para processar arquivos CSV de maneira assíncrona e confiável. Este projeto utiliza o **Kafka** como broker de mensagens e foi desenvolvido utilizando **Golang**, com o auxílio do framework **Gin**. A arquitetura da aplicação foi projetada para suportar validação, integração com Kafka, geração de boletos e envio de e-mails para notificações automáticas.

---

## ⚙️ **Pré-requisitos**

Certifique-se de ter instalado:

- **Docker** e **Docker Compose**.
- Ferramenta de teste para APIs como **Postman**, **Insomnia**, ou **cURL**.

---

## 🚀 **Instruções de Configuração e Execução**

### **1. Clonar o Repositório**

Clone o repositório do projeto para sua máquina:

```bash
git clone https://github.com/marcosedson/kanastra-api.git
cd kanastra-api
```

---

### **2. Iniciar os Contêineres**

Os arquivos `Dockerfile` e `docker-compose.yml` estão localizados na raiz do projeto. Para iniciar todos os serviços, incluindo Kafka e a aplicação, use:

```bash
docker compose up -d --build
```

Depois de iniciada, a aplicação estará disponível em:

```plaintext
http://localhost:8084
```

Você pode acompanhar os logs em tempo real usando:

```bash
docker logs -f kanastra-api
```

---

### **3. Testar o Projeto**

#### **Testes Unitários**
Executar o comando do docker antes caso não tenha executado

```bash
docker compose up -d --build
```

Execute os testes unitários do projeto com o comando:

```bash
go test ./... -v
```

---

### **4. Endpoints Disponíveis**

#### **Processar Arquivos CSV**

- **Endpoint**: `POST /process-files`
- **Descrição**: Este endpoint permite o envio de um ou mais arquivos CSV para processamento. Os dados são validados e, após isso, enviados para o Kafka.
- **Requisição**:
   - Tipo de dado: `multipart/form-data`.
   - Chave esperada: `files` com um ou mais arquivos CSV anexados.
- **Exemplo de uso (cURL)**:

```bash
curl -X POST -F 'files=@exemplo.csv' \
                -F 'files=@exemplo2.csv' \
                http://localhost:8084/process-files
```

- **Resposta**:

```json
{
  "message": "Files are being processed"
}
```

##### **Validação de Arquivos CSV**
- **Cabeçalho esperado no arquivo CSV**:
   - `name,governmentId,email,debtAmount,debtDueDate,debtId`
- Após a validação, cada linha do arquivo é enviada para o Kafka.

---

## 🛠️ **Arquitetura do Projeto**

### **1. Camadas Principais**

- **`/internal/core/domain/`**: Define as entidades principais da aplicação, como a estrutura `Debt`.
- **`/internal/core/usecase/`**: Implementa o caso de uso principal do sistema (processamento de arquivos de forma assíncrona).
- **`/internal/handler/`**: Contém os controladores responsáveis pelo recebimento das requisições HTTP.
- **`/internal/infra/adapter/`**:
   - **`external/`**: Adaptações externas (ex.: geradores de boleto, serviços de e-mail).
   - **`persistence/`**: Controla o armazenamento em memória dos arquivos já processados.
   - **`kafka/`**: Producer e consumer do Kafka.
- **`/internal/setup/`**: Arquivos para inicialização da aplicação, como dependências e rotas.
- **`/internal/infra/config/`**: Manuseio de variáveis de ambiente.

---

### **2. Estruturas Importantes**

#### **Debt**
Define um débito no sistema:

```go
type Debt struct {
    Name         string  `json:"Name"`
    GovernmentID string  `json:"GovernmentID"`
    Email        string  `json:"Email"`
    DebtAmount   float64 `json:"DebtAmount"`
    DebtDueDate  string  `json:"DebtDueDate"`
    DebtID       string  `json:"DebtID"`
}
```

#### **Fluxo de Processamento Assíncrono de Arquivos**
1. O arquivo CSV enviado pelo cliente é validado.
2. Cada linha do arquivo é analisada:
   - Validação do **GovernmentID**, **Email**, **DebtDueDate** e **DebtAmount**.
   - Caso inválida, a linha é descartada.
   - Caso válida, ela é enviada para o Kafka.
3. Durante o consumo:
   - Um boleto é gerado.
   - Um email é enviado.
   - A linha é marcada como processada no repositório.

#### **Resposta do Endpoint**
A API responde imediatamente ao cliente enquanto o processamento dos arquivos ocorre em background.

---

## 📦 **Gerenciamento de Mensagens com Kafka**

### **Tópicos Utilizados**
- **`default_topic`**:
   - Recebe cada linha processada do arquivo CSV.

### **Produtores e Consumidores**
- **Produtor (Producer)**:
   - Envia os dados do arquivo para o Kafka.
- **Consumidor (Consumer)**:
   - Processa as mensagens recebidas do Kafka. Cada mensagem é enviada para:
      - Serviço de boletos.
      - Publicador de e-mails.

---

## 📬 **Contato**

- **Autor**: Marcos Marcon
- **Email**: marcossedsonn@gmail.com
- **GitHub**: [marcosedson](https://github.com/marcosedson)
