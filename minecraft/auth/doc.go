// Package auth implements authentication to Microsoft accounts and XBOX Live accounts. It does so in a couple
// of steps, the first of which being authentication to the Live account to obtain a Live token, so that
// authentication to the XBOX Live account may be initiated.
//
// The auth package currently does not handle 2FA accounts. Trying to authenticate to an account with 2FA
// enabled will result in undefined behaviour.
package auth
