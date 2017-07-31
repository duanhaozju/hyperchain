
#ifndef HEADER_SMS4_H
#define HEADER_SMS4_H

#include <openssl/opensslconf.h>
#ifndef NO_GMSSL

#define SMS4_KEY_LENGTH		16
#define SMS4_BLOCK_SIZE		16
#define SMS4_IV_LENGTH		(SMS4_BLOCK_SIZE)
#define SMS4_NUM_ROUNDS		32

#include <sys/types.h>
#include <stdint.h>
#include <string.h>


#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
	uint32_t rk[SMS4_NUM_ROUNDS];
} sms4_key_t;

void sms4_set_encrypt_key(sms4_key_t *key, const unsigned char *user_key);
void sms4_set_decrypt_key(sms4_key_t *key, const unsigned char *user_key);
void sms4_encrypt(const unsigned char *in, unsigned char *out, const sms4_key_t *key);
#define sms4_decrypt(in,out,key)  sms4_encrypt(in,out,key)

void sms4_encrypt_init(sms4_key_t *key);
void sms4_encrypt_8blocks(const unsigned char *in, unsigned char *out, const sms4_key_t *key);
void sms4_encrypt_16blocks(const unsigned char *in, unsigned char *out, const sms4_key_t *key);

void sms4_ecb_encrypt(const unsigned char *in, unsigned char *out,
	const sms4_key_t *key, int enc);
void sms4_cbc_encrypt(const unsigned char *in, unsigned char *out,
	size_t len, const sms4_key_t *key, unsigned char *iv, int enc);
void sms4_cfb128_encrypt(const unsigned char *in, unsigned char *out,
	size_t len, const sms4_key_t *key, unsigned char *iv, int *num, int enc);
void sms4_ofb128_encrypt(const unsigned char *in, unsigned char *out,
	size_t len, const sms4_key_t *key, unsigned char *iv, int *num);
void sms4_ctr128_encrypt(const unsigned char *in, unsigned char *out,
	size_t len, const sms4_key_t *key, unsigned char *iv,
	unsigned char ecount_buf[SMS4_BLOCK_SIZE], unsigned int *num);

int sms4_wrap_key(sms4_key_t *key, const unsigned char *iv,
	unsigned char *out, const unsigned char *in, unsigned int inlen);
int sms4_unwrap_key(sms4_key_t *key, const unsigned char *iv,
	unsigned char *out, const unsigned char *in, unsigned int inlen);



#define SMS4_EDE_KEY_LENGTH	32

typedef struct {
	sms4_key_t k1;
	sms4_key_t k2;
} sms4_ede_key_t;

void sms4_ede_set_encrypt_key(sms4_ede_key_t *key, const unsigned char *user_key);
void sms4_ede_set_decrypt_key(sms4_ede_key_t *key, const unsigned char *user_key);
void sms4_ede_encrypt(sms4_ede_key_t *key, const unsigned char *in, unsigned char *out);
void sms4_ede_encrypt_8blocks(sms4_ede_key_t *key, const unsigned char *in, unsigned char *out);
void sms4_ede_encrypt_16blocks(sms4_ede_key_t *key, const unsigned char *in, unsigned char *out);
void sms4_ede_decrypt(sms4_ede_key_t *key, const unsigned char *in, unsigned char *out);
void sms4_ede_decrypt_8blocks(sms4_ede_key_t *key, const unsigned char *in, unsigned char *out);
void sms4_ede_decrypt_16blocks(sms4_ede_key_t *key, const unsigned char *in, unsigned char *out);


#ifdef __cplusplus
}
#endif
#endif
#endif
