# Korp_Teste_SeuNome

Sistema de emissão de Notas Fiscais — teste técnico Korp ERP.

## Visão geral

O sistema é composto por dois microsserviços em Go e um frontend em Angular:

- **backend/estoque** — cadastro de produtos e controle de saldo.
- **backend/faturamento** — cadastro de notas fiscais, numeração sequencial,
  status e impressão (com atualização de saldo via serviço de Estoque).
- **frontend** — aplicação Angular que consome os dois microsserviços.
