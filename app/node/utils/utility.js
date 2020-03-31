exports.getSalt = () => {
    return require('crypto').randomBytes(64).toString('hex');
};