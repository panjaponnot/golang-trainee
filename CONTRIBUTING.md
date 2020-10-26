# Contributing
1. การตั้ง API Path จะเป็นตาม Standard API ดังนี้ api/v1/... และไม่มีคำกิริยาภายในชื่อ path
2. การ Route จะมี route ใหญ่ อยู่ข้างนอก และจะแยกกันไปตามแต่ละ Feature เป็น route ย่อยต่าง ๆ
3. การตั้งชื่อ function สำหรับ api จะลงท้ายด้วย **Endpoint**
4. การตั้งชื่อตัวแปล env จะใช้ **ตัวพิมพ์ใหญ่** ทุกตัว
5. core folder จะไว้เก็บ function ต่าง ๆ ที่สามารถใช้ได้กับหลาย ๆ feature
6. pkg folder จะไว้เก็บ package ต่าง ๆ เช่น ตัว server หรือ ตัวเชื่อม Database
7. model folder จะไว้เก็บ struct สำหรับการสร้าง table ใน Database
8. การวางโค้ดจะเริ่มจาก const, public struct, private struct, public function, private function
9. การตั้งตัวแปล struct คือ capitalize, ตัวแปล คือ lower camel case
10. การกำหนดตัวแปลจาก env จะต้องเป็น **ตัวพิมพ์ใหญ่** ทั้งหมด