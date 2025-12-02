### Bulk Import Transactions

**Endpoint:**  
POST /api/transactions/bulk?workers=4  

**Request:**
```json
[
  {"category": "еда", "amount": 1200, "date": "2024-11-20"},
  {"category": "транспорт", "amount": -50, "date": "2024-11-21"}
]
