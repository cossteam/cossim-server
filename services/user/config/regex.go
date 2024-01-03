package config

const PasswordRegex = `^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])[\w@#$%^&+=!-]{8,20}$`
const EmailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
