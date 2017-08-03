#include <cstdio>
#include <cstdlib>
#include <cassert>
#include <ctime>

#include <openssl/bn.h>
#include <openssl/ec.h>
#include <openssl/evp.h>
#include <openssl/rand.h>
#include <openssl/engine.h>
#include <openssl/sm2.h>
#define rep(i,n) for (int i=0;i<n;++i)

EC_GROUP *ec_group;
EC_KEY *key;
BN_CTX* ctx;

#define CURVE NID_sm2p256v1

#define xputs2(str,name) printf(#name ": %s\n", str)
#define xputs(str) xputs2(str,str)
#define pbn(bn) xputs2(BN_bn2hex(bn),bn)
#define p(x) xputs2(x, #x)

void ppntf(const EC_POINT* p, const char* name) {
    printf("point %s\n", name);
    BIGNUM *x = BN_new(), *y = BN_new(), *z = BN_new();
    EC_POINT_get_affine_coordinates_GFp(ec_group, p, x, y, ctx);
    pbn(x); pbn(y); //pbn(z);
}
#define ppnt(x) ppntf(x, #x)

int main(int argc, const char *argv[])
{
    ctx = BN_CTX_new();
    ec_group = EC_GROUP_new_by_curve_name(CURVE); //();
    key = EC_KEY_new_by_curve_name(CURVE);//NID_sm2p256v1);
    EC_KEY_generate_key(key);
    const unsigned char* dgst = (const unsigned char*)"a567f231678689de";
    unsigned int dlen = sizeof(dgst);
    struct timespec tm, tm2;
    //ECDSA_SIG* sig = SM2_do_sign(dgst, dlen, key);
    unsigned char sig[256];
    unsigned int siglen;
    SM2_sign(NID_undef, dgst, dlen, sig, &siglen, key);

    p(dgst);
    auto priv = EC_KEY_get0_private_key(key); pbn(priv);
    auto pub = EC_KEY_get0_public_key(key); ppnt(pub);
    printf("sig(%d): ", siglen); for (int i=siglen-1;i>=0;--i) printf("%02X", sig[i]); putchar('\n');

    return 0;
}

