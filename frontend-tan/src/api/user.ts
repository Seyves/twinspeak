import { httpClient } from '@/lib/httpClient'

export type UserInfo = {
    email: string
    profilePicture: string | null
    createdAt: string
}

export type CreditGrant = {
    id: string
    amount: number
    remainingAmount: number
    type: 'monthly' | 'topup'
    expiresAt: string | null
    createdAt: string
}

export async function getMe(): Promise<UserInfo> {
    const response = await httpClient.get('me')
    return response.json()
}

export async function getCreditGrants(): Promise<CreditGrant[]> {
    const response = await httpClient.get('me/credits')
    return response.json()
}
