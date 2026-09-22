export interface User {
  id: string
  username: string
  isAdmin: boolean
  createdAt: string
  updatedAt: string
}

export interface UserPayload {
  username: string
  password: string
}
