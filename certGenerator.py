from OpenSSL import crypto
import random
import string


def cert_gen(
    emailAddress="emailAddress",
    commonName="commonName",
    countryName="NT",
    localityName="localityName",
    stateOrProvinceName="stateOrProvinceName",
    organizationName="organizationName",
    organizationUnitName="organizationUnitName",
    serialNumber=0,
    validityStartInSeconds=0,
    validityEndInSeconds=10*365*24*60*60,
    KEY_FILE = "private.key",
    CERT_FILE="selfsigned.crt", k=any):
    #can look at generated file using openssl:
    #openssl x509 -inform pem -in selfsigned.crt -noout -text
    # create a key pair
   
    # create a self-signed cert
    cert = crypto.X509()



    cert.get_subject().C = countryName
    cert.get_subject().ST = stateOrProvinceName
    cert.get_subject().L = localityName
    cert.get_subject().O = organizationName
    cert.get_subject().OU = organizationUnitName
    cert.get_subject().CN = commonName
    cert.get_subject().emailAddress = emailAddress
    cert.set_serial_number(serialNumber)
    cert.gmtime_adj_notBefore(validityStartInSeconds)
    cert.gmtime_adj_notAfter(validityEndInSeconds)
    cert.set_issuer(cert.get_subject())
    cert.set_pubkey(k)
    cert.sign(k, 'sha512')
    return crypto.dump_certificate(crypto.FILETYPE_PEM, cert).decode("utf-8")
    #with open(CERT_FILE, "wt") as f:
    #    f.write(crypto.dump_certificate(crypto.FILETYPE_PEM, cert).decode("utf-8"))
    #with open(KEY_FILE, "wt") as f:
    #    f.write(crypto.dump_privatekey(crypto.FILETYPE_PEM, k).decode("utf-8"))

k = crypto.PKey()
k.generate_key(crypto.TYPE_RSA, 4096)
certs = ""
certsToGen = 20000
letters = string.ascii_lowercase

for j in range(1,6):
    certs = ""
    print("working on file: ", j)
    for i in range(certsToGen):
        if i % 500 ==0:
            k.generate_key(crypto.TYPE_RSA, 4096)
        cName = str(random.randint(10,99))
        stName = str(random.randint(10,99))# write_random_lowercase(10).decode("utf-8")# ''.join(random.choice(letters) for i in range(10))
        lName =str(random.randint(10,99))#write_random_lowercase(10).decode("utf-8")# ''.join(random.choice(letters) for i in range(10))
        oName =str(random.randint(10,99))# write_random_lowercase(10).decode("utf-8")#''.join(random.choice(letters) for i in range(10))
        ouName =str(random.randint(10,99))#write_random_lowercase(10).decode("utf-8")# ''.join(random.choice(letters) for i in range(10))
        cnName = str(random.randint(10,99))#write_random_lowercase(10).decode("utf-8")#''.join(random.choice(letters) for i in range(10))
        eAddress =str(random.randint(10,99))# write_random_lowercase(10).decode("utf-8")#''.join(random.choice(letters) for i in range(20))
        certs = certs + cert_gen(countryName=cName, stateOrProvinceName=stName, localityName=lName, organizationName=oName, organizationUnitName=ouName, commonName=cnName, emailAddress=eAddress, k=k)


    CERT_FILE="AllCertsOneFile20000-"+str(j)+".crt"

    with open(CERT_FILE, "wt") as f:
        f.write(certs)
