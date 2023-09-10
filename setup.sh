DIR=$(dirname $0)
echo "directory: ${DIR}"

USER="root"
echo "user: ${USER}"

mysql -u ${USER} -p < ${DIR}/db.sql