// \file:sm2.c
//SM2 Algorithm
//2011-11-10
//author:goldboar
//email:goldboar@163.com
//depending:opnessl library

//SM2 Standards: http://www.oscca.gov.cn/News/201012/News_1197.htm
#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
#include "./openssl/ec.h"
#include "./openssl/rand.h"
#include "./openssl/err.h"
#include "./openssl/obj_mac.h"
#include "./openssl/ecdsa.h"
#include "./openssl/ecdh.h"
#include "./openssl/bn.h"
#include "./openssl/evp.h"
#include "./openssl/engine.h"
#include "sm2.h"
#include "./openssl/kdf.h"

#define  NID_X9_62_prime_field 406

#ifdef OPENSSL_NO_EC2M
static int abort_no_GF2m() { 
    fputs("Fail : no GF2m support, please re-compile program without OPENSSL_NO_EC2M defined. "
            "see crypto/guomi/crypto/sm2 & crypto/guomi/gm_sm2.go. Contact Xu Qiyuan (x@xeroe.net or tel: +8618069862522) in any problem.\n", 
            stderr); abort(); return 0; 
}
#define EC_POINT_get_affine_coordinates_GF2m(p1,p2,p3,p4,p5) abort_no_GF2m()
#endif



static void BNPrintf(BIGNUM* bn)
{
	char *p=NULL;
	p=BN_bn2hex(bn);
	printf("%s",p);
	OPENSSL_free(p);
}


static int sm2_sign_setup(EC_KEY *eckey, BN_CTX *ctx_in, BIGNUM **kp, BIGNUM **rp)
{
	BN_CTX   *ctx = NULL;
	BIGNUM	 *k = NULL, *r = NULL, *order = NULL, *X = NULL;
	EC_POINT *tmp_point=NULL;
	const EC_GROUP *group;
	int 	 ret = 0;

	if (eckey == NULL || (group = EC_KEY_get0_group(eckey)) == NULL)
	{
		ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_PASSED_NULL_PARAMETER);
		return 0;
	}

	if (ctx_in == NULL) 
	{
		if ((ctx = BN_CTX_new()) == NULL)
		{
			ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP,ERR_R_MALLOC_FAILURE);
			return 0;
		}
	}
	else
		ctx = ctx_in;

	k     = BN_new();	/* this value is later returned in *kp */
	r     = BN_new();	/* this value is later returned in *rp */
	order = BN_new();
	X     = BN_new();
	if (!k || !r || !order || !X)
	{
		ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_MALLOC_FAILURE);
		goto err;
	}
	if ((tmp_point = EC_POINT_new(group)) == NULL)
	{
		ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_EC_LIB);
		goto err;
	}
	if (!EC_GROUP_get_order(group, order, ctx))
	{
		ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_EC_LIB);
		goto err;
	}
	
	do
	{
		/* get random k */	
		do
			if (!BN_rand_range(k, order))
			{
				ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ECDSA_R_RANDOM_NUMBER_GENERATION_FAILED);	
				goto err;
			}
		while (BN_is_zero(k));

		/* compute r the x-coordinate of generator * k */
		if (!EC_POINT_mul(group, tmp_point, k, NULL, NULL, ctx))
		{
			ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_EC_LIB);
			goto err;
		}
		if (EC_METHOD_get_field_type(EC_GROUP_method_of(group)) == NID_X9_62_prime_field)
		{
			if (!EC_POINT_get_affine_coordinates_GFp(group,
				tmp_point, X, NULL, ctx))
			{
				ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP,ERR_R_EC_LIB);
				goto err;
			}
		}
		else /* NID_X9_62_characteristic_two_field */
		{
			if (!EC_POINT_get_affine_coordinates_GF2m(group,
				tmp_point, X, NULL, ctx))
			{
				ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP,ERR_R_EC_LIB);
				goto err;
			}
		}
		if (!BN_nnmod(r, X, order, ctx))
		{
			ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_BN_LIB);
			goto err;
		}
	}
	while (BN_is_zero(r));

	/* compute the inverse of k */
// 	if (!BN_mod_inverse(k, k, order, ctx))
// 	{
// 		ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_BN_LIB);
// 		goto err;	
// 	}
	/* clear old values if necessary */
	if (*rp != NULL)
		BN_clear_free(*rp);
	if (*kp != NULL) 
		BN_clear_free(*kp);
	/* save the pre-computed values  */
	*rp = r;
	*kp = k;
	ret = 1;
err:
	if (!ret)
	{
		if (k != NULL) BN_clear_free(k);
		if (r != NULL) BN_clear_free(r);
	}
	if (ctx_in == NULL) 
		BN_CTX_free(ctx);
	if (order != NULL)
		BN_free(order);
	if (tmp_point != NULL) 
		EC_POINT_free(tmp_point);
	if (X)
		BN_clear_free(X);
	return(ret);
}


static ECDSA_SIG *sm2_do_sign(const unsigned char *dgst, int dgst_len, const BIGNUM *in_k, const BIGNUM *in_r, EC_KEY *eckey)
{
	int     ok = 0, i;
	BIGNUM *k=NULL, *s, *m=NULL,*tmp=NULL,*order=NULL;
	const BIGNUM *ck;
	BN_CTX     *ctx = NULL;
	const EC_GROUP   *group;
	ECDSA_SIG  *ret;
	//ECDSA_DATA *ecdsa;
	const BIGNUM *priv_key;
    BIGNUM *r,*x=NULL,*a=NULL;	//new added
	//ecdsa    = ecdsa_check(eckey);
	group    = EC_KEY_get0_group(eckey);
	priv_key = EC_KEY_get0_private_key(eckey);
	
	if (group == NULL || priv_key == NULL /*|| ecdsa == NULL*/)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_PASSED_NULL_PARAMETER);
		return NULL;
	}

	ret = ECDSA_SIG_new();
	if (!ret)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_MALLOC_FAILURE);
		return NULL;
	}
	s = ret->s;
	r = ret->r;

	if ((ctx = BN_CTX_new()) == NULL || (order = BN_new()) == NULL ||
		(tmp = BN_new()) == NULL || (m = BN_new()) == NULL || 
		(x = BN_new()) == NULL || (a = BN_new()) == NULL)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_MALLOC_FAILURE);
		goto err;
	}

	if (!EC_GROUP_get_order(group, order, ctx))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_EC_LIB);
		goto err;
	}
// 	for(i=0;i<dgst_len;i++)
// 		printf("%02X",dgst[i]);
//  	printf("\n");
	i = BN_num_bits(order);
	/* Need to truncate digest if it is too long: first truncate whole
	 * bytes.
	 */
	if (8 * dgst_len > i)
		dgst_len = (i + 7)/8;
	if (!BN_bin2bn(dgst, dgst_len, m))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_BN_LIB);
		goto err;
	}
	/* If still too long truncate remaining bits with a shift */
	if ((8 * dgst_len > i) && !BN_rshift(m, m, 8 - (i & 0x7)))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_BN_LIB);
		goto err;
	}
// 	fprintf(stdout,"m: ");
// 	BNPrintf(m);
// 	fprintf(stdout,"\n");
	do
	{
		if (in_k == NULL || in_r == NULL)
		{
			if (!sm2_sign_setup(eckey, ctx, &k, &x))
			{
				ECDSAerr(ECDSA_F_ECDSA_DO_SIGN,ERR_R_ECDSA_LIB);
				goto err;
			}
			ck = k;
		}
		else
		{
			ck  = in_k;
			if (BN_copy(x, in_r) == NULL)
			{
				ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_MALLOC_FAILURE);
				goto err;
			}
		}
		
		//r=(e+x1) mod n
		if (!BN_mod_add_quick(r, m, x, order))
		{
			ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_BN_LIB);
			goto err;
		}

// 	    BNPrintf(r);
// 		fprintf(stdout,"\n");

		if(BN_is_zero(r) )
			continue;

		BN_add(tmp,r,ck);
		if(BN_ucmp(tmp,order) == 0)
			continue;
		
		if (!BN_mod_mul(tmp, priv_key, r, order, ctx))
		{
			ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_BN_LIB);
			goto err;
		}
		if (!BN_mod_sub_quick(s, ck, tmp, order))
		{
			ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_BN_LIB);
			goto err;
		}
		BN_one(a);
		//BN_set_word((a),1);

		if (!BN_mod_add_quick(tmp, priv_key, a, order))
		{
			ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_BN_LIB);
			goto err;
		}
		/* compute the inverse of 1+dA */
		if (!BN_mod_inverse(tmp, tmp, order, ctx))
		{
			ECDSAerr(ECDSA_F_ECDSA_SIGN_SETUP, ERR_R_BN_LIB);
			goto err;	
		}
// 		BNPrintf(tmp);
// 		fprintf(stdout,"\n");

		if (!BN_mod_mul(s, s, tmp, order, ctx))
		{
			ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_BN_LIB);
			goto err;
		}
		if (BN_is_zero(s))
		{
			/* if k and r have been supplied by the caller
			 * don't to generate new k and r values */
			if (in_k != NULL && in_r != NULL)
			{
				ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ECDSA_R_NEED_NEW_SETUP_VALUES);
				goto err;
			}
		}
		else
			/* s != 0 => we have a valid signature */
			break;
	}
	while (1);

	ok = 1;
err:
	if (!ok)
	{
		ECDSA_SIG_free(ret);
		ret = NULL;
	}
	if (ctx)
		BN_CTX_free(ctx);
	if (m)
		BN_clear_free(m);
	if (tmp)
		BN_clear_free(tmp);
	if (order)
		BN_free(order);
	if (k)
		BN_clear_free(k);
	if (x)
		BN_clear_free(x);
	if (a)
		BN_clear_free(a);
	return ret;
}

static int sm2_do_verify(const unsigned char *dgst, int dgst_len,
		const ECDSA_SIG *sig, EC_KEY *eckey)
{
	int ret = -1, i;
	BN_CTX   *ctx;
	BIGNUM   *order, *R,  *m, *X,*t;
	EC_POINT *point = NULL;
	const EC_GROUP *group;
	const EC_POINT *pub_key;

	/* check input values */
	if (eckey == NULL || (group = EC_KEY_get0_group(eckey)) == NULL ||
	    (pub_key = EC_KEY_get0_public_key(eckey)) == NULL || sig == NULL)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ECDSA_R_MISSING_PARAMETERS);
		return -1;
	}

	ctx = BN_CTX_new();
	if (!ctx)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_MALLOC_FAILURE);
		return -1;
	}
	BN_CTX_start(ctx);
	order = BN_CTX_get(ctx);	
	R    = BN_CTX_get(ctx);
	t    = BN_CTX_get(ctx);
	m     = BN_CTX_get(ctx);
	X     = BN_CTX_get(ctx);
	if (!X)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_BN_LIB);
		goto err;
	}
	
	if (!EC_GROUP_get_order(group, order, ctx))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_EC_LIB);
		goto err;
	}

	if (BN_is_zero(sig->r)          || BN_is_negative(sig->r) || 
	    BN_ucmp(sig->r, order) >= 0 || BN_is_zero(sig->s)  ||
	    BN_is_negative(sig->s)      || BN_ucmp(sig->s, order) >= 0)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ECDSA_R_BAD_SIGNATURE);
		ret = 0;	/* signature is invalid */
		goto err;
	}

	//t =(r+s) mod n
	if (!BN_mod_add_quick(t, sig->s, sig->r,order))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_BN_LIB);
		goto err;
	}
	if (BN_is_zero(t))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ECDSA_R_BAD_SIGNATURE);
		ret = 0;	/* signature is invalid */
		goto err;
	}
	
	//point = s*G+t*PA
	if ((point = EC_POINT_new(group)) == NULL)
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_MALLOC_FAILURE);
		goto err;
	}
	if (!EC_POINT_mul(group, point, sig->s, pub_key, t, ctx))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_EC_LIB);
		goto err;
	}
	if (EC_METHOD_get_field_type(EC_GROUP_method_of(group)) == NID_X9_62_prime_field)
	{
		if (!EC_POINT_get_affine_coordinates_GFp(group,
			point, X, NULL, ctx))
		{
			ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_EC_LIB);
			goto err;
		}
	}
	else /* NID_X9_62_characteristic_two_field */
	{
		if (!EC_POINT_get_affine_coordinates_GF2m(group,
			point, X, NULL, ctx))
		{
			ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_EC_LIB);
			goto err;
		}
	}
 	
	i = BN_num_bits(order);
	/* Need to truncate digest if it is too long: first truncate whole
	 * bytes.
	 */
	if (8 * dgst_len > i)
		dgst_len = (i + 7)/8;
	if (!BN_bin2bn(dgst, dgst_len, m))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_BN_LIB);
		goto err;
	}
	/* If still too long truncate remaining bits with a shift */
	if ((8 * dgst_len > i) && !BN_rshift(m, m, 8 - (i & 0x7)))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_BN_LIB);
		goto err;
	}

	/* R = m + X mod order */
	if (!BN_mod_add_quick(R, m, X, order))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_VERIFY, ERR_R_BN_LIB);
		goto err;
	}

	/*  if the signature is correct R is equal to sig->r */
	ret = (BN_ucmp(R, sig->r) == 0);
err:
	BN_CTX_end(ctx);
	BN_CTX_free(ctx);
	if (point)
		EC_POINT_free(point);
	return ret;
}


EC_POINT *sm2_compute_key(const EC_POINT *b_pub_key_r, const EC_POINT *b_pub_key, const BIGNUM *a_r,EC_KEY *a_eckey)
{
	BN_CTX *ctx;
	EC_POINT *tmp=NULL;
	BIGNUM *x=NULL, *y=NULL, *order=NULL,*z=NULL;
	const BIGNUM *priv_key;
	const EC_GROUP* group;
	EC_POINT *ret= NULL;
/*	size_t buflen, len;*/
	unsigned char *buf=NULL;
	int i, j;
	//char *p=NULL;
	BIGNUM *x1,*x2,*t,*h;

	if ((ctx = BN_CTX_new()) == NULL) goto err;
	BN_CTX_start(ctx);
	x = BN_CTX_get(ctx);
	y = BN_CTX_get(ctx);
	order = BN_CTX_get(ctx);
	z = BN_CTX_get(ctx);
	x1 = BN_CTX_get(ctx);
	x2 = BN_CTX_get(ctx);
	t = BN_CTX_get(ctx);
	h = BN_CTX_get(ctx);

	
	priv_key = EC_KEY_get0_private_key(a_eckey);
	if (priv_key == NULL)
	{
		ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_NO_PRIVATE_VALUE);
		goto err;
	}

	group = EC_KEY_get0_group(a_eckey);
	if ((tmp=EC_POINT_new(group)) == NULL)
	{
		ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ERR_R_MALLOC_FAILURE);
		goto err;
	}

	if (!EC_POINT_mul(group, tmp, a_r, NULL, NULL, ctx)) 
	{
		ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
		goto err;
	}
	
	if (EC_METHOD_get_field_type(EC_GROUP_method_of(group)) == NID_X9_62_prime_field) 
	{
		if (!EC_POINT_get_affine_coordinates_GFp(group, tmp, x, NULL, ctx)) 
		{
			ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
			goto err;
		}
	}
	else
	{
		if (!EC_POINT_get_affine_coordinates_GF2m(group, tmp, x, NULL, ctx)) 
		{
			ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
			goto err;
		}
	}
	
	if (!EC_GROUP_get_order(group, order, ctx))
	{
		ECDSAerr(ECDSA_F_ECDSA_DO_SIGN, ERR_R_EC_LIB);
		goto err;
	}
		
	i = BN_num_bits(order);
	j = i/2 -1;
	BN_mask_bits(x,j);
	BN_set_word(y,2);
	BN_set_word(z,j);
	BN_exp(y,y,z,ctx);
	BN_add(x1,x,y);
	
// 	fprintf(stdout,"X1=: ");
// 	BNPrintf(x1);
// 	fprintf(stdout,"\n");

	BN_mod_mul(t,x1,a_r,order,ctx);
	BN_mod_add_quick(t,t,priv_key,order);
// 
// 	fprintf(stdout,"ta=: ");
// 	BNPrintf(t);
// 	fprintf(stdout,"\n");

	
	if (EC_METHOD_get_field_type(EC_GROUP_method_of(group)) == NID_X9_62_prime_field) 
	{
		if (!EC_POINT_get_affine_coordinates_GFp(group, b_pub_key_r, x, NULL, ctx)) 
		{
			ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
			goto err;
		}
	}
	else
	{
		if (!EC_POINT_get_affine_coordinates_GF2m(group, b_pub_key_r, x, NULL, ctx)) 
		{
			ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
			goto err;
		}
	}

	i = BN_num_bits(order);
	j = i/2 -1;
	BN_mask_bits(x,j);
	BN_set_word(y,2);
	BN_set_word(z,j);
	BN_exp(y,y,z,ctx);
	BN_add(x2,x,y);
	
// 	fprintf(stdout,"X2=: ");
// 	BNPrintf(x2);
// 	fprintf(stdout,"\n");


	//x2*Rb+Pb;
	if (!EC_POINT_mul(group, tmp, NULL,b_pub_key_r,x2,ctx) )
	{
		ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
		goto err;
	}
	if ((ret=EC_POINT_new(group)) == NULL)
	{
		ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ERR_R_MALLOC_FAILURE);
		goto err;
	}
	if (!EC_POINT_add(group, ret, b_pub_key, tmp, ctx))
	{
		ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
		goto err;
	}
	if (!EC_POINT_get_affine_coordinates_GFp(group,ret, x, y, ctx)) 
	{
		goto err;
	}
// 	fprintf(stdout, "\nTesting x2*Rb+Pb Key Point\n     x = 0x");
// 	BNPrintf(x);
// 	fprintf(stdout, "\n     y = 0x");
// 	BNPrintf( y);
// 	fprintf(stdout, "\n");
// 	
	if(!EC_GROUP_get_cofactor(group, h, ctx))
	{
		goto err;
	}
    BN_mul(t,t,h,ctx);

	//h*t*(x2*Rb+Pb)
	if (!EC_POINT_mul(group, ret, NULL,ret,t,ctx) ) 
	{
		goto err;
	}
	if (!EC_POINT_get_affine_coordinates_GFp(group,ret, x, y, ctx)) 
	{
		goto err;
	}
// 	fprintf(stdout, "\nTesting ret Key Point\n     x = 0x");
// 	BNPrintf(x);
// 	fprintf(stdout, "\n     y = 0x");
// 	BNPrintf( y);
// 	fprintf(stdout, "\n");

	
err:
	if (tmp) EC_POINT_free(tmp);
	if (ctx) BN_CTX_end(ctx);
	if (ctx) BN_CTX_free(ctx);
	if (buf) OPENSSL_free(buf);
	return(ret);
}

/** SM2_sign_setup
* precompute parts of the signing operation. 
* \param eckey pointer to the EC_KEY object containing a private EC key
* \param ctx  pointer to a BN_CTX object (may be NULL)
* \param k pointer to a BIGNUM pointer for the inverse of k
* \param rp   pointer to a BIGNUM pointer for x coordinate of k * generator
* \return 1 on success and 0 otherwise
 */

int  SM2_sign_setup(EC_KEY *eckey, BN_CTX *ctx_in, BIGNUM **kinvp, BIGNUM **rp)
{
// 	ECDSA_DATA *ecdsa = ecdsa_check(eckey);
// 	if (ecdsa == NULL)
// 		return 0;
	return SM2_sign_setup(eckey, ctx_in, kinvp, rp); 
}
/** SM2_sign_ex
 * computes ECDSA signature of a given hash value using the supplied
 * private key (note: sig must point to ECDSA_size(eckey) bytes of memory).
 * \param type this parameter is ignored
 * \param dgst pointer to the hash value to sign
 * \param dgstlen length of the hash value
 * \param sig buffer to hold the DER encoded signature
 * \param siglen pointer to the length of the returned signature
 * \param k optional pointer to a pre-computed inverse k
 * \param rp optional pointer to the pre-computed rp value (see 
 *        ECDSA_sign_setup
 * \param eckey pointer to the EC_KEY object containing a private EC key
 * \return 1 on success and 0 otherwise
 */
int	  SM2_sign_ex(int type, const unsigned char *dgst, int dlen, unsigned char 
	*sig, unsigned int *siglen, const BIGNUM *kinv, const BIGNUM *r, 
	EC_KEY *eckey)
{
	ECDSA_SIG *s;
	RAND_seed(dgst, dlen);
	s = sm2_do_sign(dgst, dlen, kinv, r, eckey);
	if (s == NULL)
	{
		*siglen=0;
		return 0;
	}
	*siglen = i2d_ECDSA_SIG(s, &sig);
	ECDSA_SIG_free(s);
	return 1;
}

/** SM2_sign
  * computes ECDSA signature of a given hash value using the supplied
  * private key (note: sig must point to ECDSA_size(eckey) bytes of memory).
  * \param type this parameter is ignored
  * \param dgst pointer to the hash value to sign
  * \param dgstlen length of the hash value
  * \param sig buffer to hold the DER encoded signature
  * \param siglen pointer to the length of the returned signature
  * \param eckey pointer to the EC_KEY object containing a private EC key
  * \return 1 on success and 0 otherwise
 */
int	  SM2_sign(int type, const unsigned char *dgst, int dlen, unsigned char 
		*sig, unsigned int *siglen, EC_KEY *eckey)
{

	return SM2_sign_ex(type, dgst, dlen, sig, siglen, NULL, NULL, eckey);

}


/** SM2_verify
  * verifies that the given signature is valid ECDSA signature
  * of the supplied hash value using the specified public key.
  * \param type this parameter is ignored
  * \param dgst pointer to the hash value 
  * \param dgstlen length of the hash value
  * \param sig  pointer to the DER encoded signature
  * \param siglen length of the DER encoded signature
  * \param eckey pointer to the EC_KEY object containing a public EC key
  * \return 1 if the signature is valid, 0 if the signature is invalid and -1 on error
  */
int SM2_verify(int type, const unsigned char *dgst, int dgst_len,
		const unsigned char *sigbuf, int sig_len, EC_KEY *eckey)
 {
	ECDSA_SIG *s;
	int ret=-1;

	s = ECDSA_SIG_new();
	if (s == NULL) return(ret);
	if (d2i_ECDSA_SIG(&s, &sigbuf, sig_len) == NULL) goto err;
	ret=sm2_do_verify(dgst, dgst_len, s, eckey);
err:
	ECDSA_SIG_free(s);
	return(ret);
}

int SM2_DH_key(const EC_GROUP * group, const EC_POINT *b_pub_key_r, const EC_POINT *b_pub_key, const BIGNUM *a_r,EC_KEY *a_eckey,
			   unsigned char *outkey,size_t keylen)
{
	EC_POINT *dhpoint = NULL;
	BN_CTX * ctx;
	EC_POINT *P;
	BIGNUM *x, *y;
	int ret = 0;
	unsigned char in[128];
	int inlen;
	int len;

	P = EC_POINT_new(group);
	if (!P ) goto err;
	ctx = BN_CTX_new();
	x = BN_new();
	y = BN_new();
	if (!x || !y ) goto err;
	
	dhpoint = sm2_compute_key(b_pub_key_r,b_pub_key,a_r,a_eckey);

	if (EC_METHOD_get_field_type(EC_GROUP_method_of(group)) == NID_X9_62_prime_field) 
	{
		if (!EC_POINT_get_affine_coordinates_GFp(group,dhpoint, x, y, ctx))
		{
			fprintf(stdout, " failed\n");
			goto err;
		}
	}
	else
	{
		if (!EC_POINT_get_affine_coordinates_GF2m(group,dhpoint, x, y, ctx)) 
		{
			ECDHerr(ECDH_F_ECDH_COMPUTE_KEY,ECDH_R_POINT_ARITHMETIC_FAILURE);
			goto err;
		}
	}

// 	if (!EC_POINT_get_affine_coordinates_GFp(group,dhpoint, x, y, ctx))
// 	{
// 		fprintf(stdout, " failed\n");
// 		goto err;
// 	}
	fprintf(stdout, "\nTesting DH Point\n     Xv = 0x");
	BNPrintf(x);
	fprintf(stdout, "\n     Yv = 0x");
	BNPrintf( y);
	fprintf(stdout, "\n");

	len = BN_bn2bin(x,in);
	inlen =BN_bn2bin(y,in+len);
	inlen = inlen + len;
	ret = x9_63_kdf(EVP_sha256(),in,inlen,keylen,outkey);
	//ret  = 1;
err:
	EC_POINT_free(P);
	EC_POINT_free(dhpoint);
	BN_CTX_free(ctx);

	return ret;
}
// ----- KDF FUNCTIONS END -----
#define xck(x) if (x != 1) return 0;


// e 是此签名所签署的 Hash，如果不造怎么生成，请看 GMssL 源代码 crypto/sm2/sm2_sign.c(ver 1.2.1):405,
// 这个 e 就是 sm2_do_verify 函数中的 e
// BN_bin2bn(dgst, dgstlen, e)
// ybit : pubkey.y 的最低位
// ctx : optional 可以为 NULL，不过以前测试时有时传 NULL 会炸，最好用 BN_CTX_new() 生成个传过来。据说全局用同一的 BN_CTX 会加速一点 BIGNUM 运算？
// return val: 遵守 OpenSSL, 1 表示成功，0为失败
//int recovery_pubkey(const EC_GROUP* ec_group, const ECDSA_SIG* sig, const unsigned char* dgst, const size_t dgst_len,
//        const char ybit, EC_POINT* pubkey, BN_CTX* ctx) {
//
//    BIGNUM* e = BN_bin2bn(dgst, dgst_len, NULL);
//
//    const BIGNUM *r, *s;
//    ECDSA_SIG_get0(sig, &r, &s);
//    static const BIGNUM *Prm = NULL, *Order = NULL;
//    if (!Prm) {
//        BIGNUM *t1 = BN_new(), *t2 = BN_new(), *tp = BN_new();
//        xck(EC_GROUP_get_curve_GFp(ec_group, tp, t1, t2, ctx));
//        Prm = tp;
//        BN_free(t1); BN_free(t2);
//    }
//    if (!Order) {
//        BIGNUM* t1 = BN_new();
//        xck(EC_GROUP_get_order(ec_group, t1, ctx));
//        Order = t1;
//    }
//    BIGNUM *xk = BN_new(), *ras = BN_new(), *rss = BN_new(),
//           *zero = BN_new();
//    BN_mod_add(ras, r, s, Order, ctx);
//    BN_mod_inverse(ras, ras, Order, ctx);
//    BN_mod_sub(rss, zero, s, Order, ctx);
//    BN_mod_mul(rss, rss, ras, Order, ctx);
//    BN_mod_sub(xk, r, e, Order, ctx);
//
//    EC_POINT *kpnt = EC_POINT_new(ec_group);
//    xck(EC_POINT_set_compressed_coordinates_GFp(ec_group, kpnt, xk, ybit, ctx));
//    //xck(EC_POINT_get_affine_coordinates_GFp(ec_group, kpnt, pbx, pby, ctx));
//    //printf("xx: %d, ybit: %d\n", BN_is_bit_set(pby, 0), ybit);
//    xck(EC_POINT_mul(ec_group, pubkey, rss, kpnt, ras, ctx));
//    return 1;
//}

EC_GROUP* sm2_ec_group() {
    static EC_GROUP* ecgroup = NULL;
    if (ecgroup) return ecgroup;
    EC_GROUP* ecg = EC_GROUP_new(EC_GFp_mont_method());
    BN_CTX* ctx = BN_CTX_new();
    BIGNUM *p = NULL, *a = NULL, *b = NULL;
    BN_hex2bn(&p, "FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF");
    BN_hex2bn(&a, "FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFC");
    BN_hex2bn(&b, "28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93");
    xck(EC_GROUP_set_curve_GFp(ecg, p, a, b, ctx));
    BIGNUM *Gx = NULL, *Gy = NULL, *N = NULL;
    BN_hex2bn(&Gx, "32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7");
    BN_hex2bn(&Gy, "BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0");
    BN_hex2bn(&N , "FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123");
    EC_POINT* generator = EC_POINT_new(ecg);
    EC_POINT_set_affine_coordinates_GFp(ecg, generator, Gx, Gy, ctx);
    EC_GROUP_set_generator(ecg, generator, N, BN_value_one());
    BN_free(p);
    BN_free(a);
    BN_free(b);
    BN_free(Gx);
    BN_free(Gy);
    BN_free(N);
    BN_CTX_free(ctx);
    return ecgroup = ecg;
}
//int recovery_pubkey_silly(const ECDSA_SIG* sig, const unsigned char* dgst, const char ybit, EC_POINT* pubkey, BN_CTX* ctx) {
//    size_t dlen = strlen((const char*)dgst);
//    return recovery_pubkey(sm2_ec_group(), sig, dgst, dlen, ybit, pubkey, ctx);
//}
//
//
//ECDSA_SIG *SM2_do_sign_compressed(const unsigned char *dgst, int dgstlen, char* ybit, EC_KEY *ec_key) {
//
//    BIGNUM *inx = BN_new(), *iny = BN_new(), *order = BN_new(), *k = BN_new();
//    BN_CTX* ctx = BN_CTX_new();
//	const EC_GROUP *ec_group = EC_KEY_get0_group(ec_key);
//    xck(EC_GROUP_get_order(ec_group, order, ctx));
//    EC_POINT *point = EC_POINT_new(ec_group);
//    do {
//        xck(BN_rand_range(k, order));
//        xck(EC_POINT_mul(ec_group, point, k, NULL, NULL, ctx));
//        xck(EC_POINT_get_affine_coordinates_GFp(ec_group, point, inx, iny, ctx));
//        BN_nnmod(inx, inx, order, ctx);
//    } while (BN_is_zero(inx));
//    BN_CTX_free(ctx);
//    *ybit = (char)!!BN_is_bit_set(iny, 0);
//    return SM2_do_sign_ex(dgst, dgstlen, k, inx, ec_key);
//}
//
//
//int SM2_sign_compressed(int type, const unsigned char *dgst, int dgstlen,
//	unsigned char *sig, unsigned int *siglen, char* ybit, EC_KEY *ec_key)
//{
//	ECDSA_SIG *s;
//
//	RAND_seed(dgst, dgstlen);
//
//	if (!(s = SM2_do_sign_compressed(dgst, dgstlen, ybit, ec_key))) {
//		*siglen = 0;
//		return 0;
//	}
//
//	*siglen = i2d_ECDSA_SIG(s, &sig);
//	ECDSA_SIG_free(s);
//
//	return 1;
//}
//
//// following is test usage
//#define xputs2(str,name) printf(#name ": %s\n", str)
//#define xputs(str) xputs2(str,str)
//#define pbn(bn) xputs2(BN_bn2hex(bn),bn)
//
//#define ppnt(x) ppntf(x, #x)
//
//EC_GROUP* ec_group;
//BN_CTX* ctx;
//
//void ppntf(const EC_POINT* p, const char* name) {
//    printf("point %s:\n", name);
//    BIGNUM *x = BN_new(), *y = BN_new(), *z = BN_new();
//    EC_POINT_get_affine_coordinates_GFp(ec_group, p, x, y, ctx);
//    pbn(x); pbn(y); //pbn(z);
//}

EC_KEY *EC_KEY_new_group()
{
    EC_KEY *ret = EC_KEY_new();
    if (ret == NULL)
        return NULL;
    EC_GROUP* group = sm2_ec_group();
    EC_KEY_set_group(ret, group);
    return ret;
}


//int main(int argc, const char *argv[])
//{
//    ctx = BN_CTX_new();
//    ec_group = EC_GROUP_new_by_curve_name(NID_sm2p256v1);
//
//    EC_KEY* key = EC_KEY_new_by_curve_name(NID_sm2p256v1);
//
//    for (int _i = 0;_i<10000;++_i) {
//    EC_KEY_generate_key(key);
//    const EC_POINT* pub = EC_KEY_get0_public_key(key);
//
//    const unsigned char* dgst = (const unsigned char*)"XERO XQY XQY XERO";
//    unsigned int dlen = strlen((const char*)dgst);
//
//    char ybit = 0;
//    ECDSA_SIG* sig = SM2_do_sign_compressed(dgst, dlen, &ybit, key);
//
//    EC_POINT* rec_pub = EC_POINT_new(ec_group);
//    int ret = recovery_pubkey_silly(sig, dgst, ybit, rec_pub, ctx);
//
//    if (!ret || EC_POINT_cmp(ec_group, pub, rec_pub, ctx)) {
//        puts("fail");
//        printf("%d\n", ret);
//        ppnt(pub);
//        ppnt(rec_pub);
//    }
//
//    }
//    return 0;
//}
