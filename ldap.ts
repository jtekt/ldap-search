import ldap from "ldapjs"
import dotenv from "dotenv"

dotenv.config()

const {
  LDAP_HOSTNAME,
  LDAP_SEARCH_OU = "OU=COMPANY,DC=COMPANY,DC=local",
  LDAP_USERNAME = "username",
  LDAP_PASSWORD = "password",
} = process.env

const connect = (): Promise<ldap.Client> =>
  new Promise((resolve, reject) => {
    const connection_options = { url: `ldap://${LDAP_HOSTNAME}` }

    const client = ldap.createClient(connection_options)

    client.on("error", (err: any) => {
      console.log(`[LDAP] Connection to ${LDAP_HOSTNAME} failed`)
      reject(err)
    })

    client.on("connect", () => {
      console.log(`[LDAP] Connection to ${LDAP_HOSTNAME} successful`)
      resolve(client)
    })

    client.on("destroy", () => {
      console.log(`[LDAP] Client destroyed`)
    })

    client.on("end", () => {
      console.log(`[LDAP] Client ended`)
    })
  })

const bind = (client: ldap.Client) =>
  new Promise((resolve, reject) => {
    client.bind(LDAP_USERNAME, LDAP_PASSWORD, (err: any) => {
      if (err) {
        console.log(`[LDAP] Bind failed`)
        reject(err)
      } else {
        console.log(`[LDAP] Bind successful`)
        resolve("")
      }
    })
  })

const unbind = (client: ldap.Client) =>
  new Promise((resolve, reject) => {
    client.unbind((err) => {
      if (err) {
        console.log(`[LDAP] Unbind failed`)
        reject(err)
      } else {
        console.log(`[LDAP] Unbind successful`)
        resolve("")
      }
    })
  })

export const search = (client: ldap.Client, search_opts: any) =>
  new Promise((resolve, reject) => {
    client.search(LDAP_SEARCH_OU, search_opts, (err, res) => {
      if (err) reject(err)
      else resolve(res)
    })
  })

const objectify_entry = (entry: any) =>
  entry.attributes.reduce(
    (acc: any, item: ldap.Attribute) => ({
      ...acc,
      [item.type]: item.values.toString(),
    }),
    {}
  )

export const handle_search_response = (res: any) =>
  new Promise((resolve, reject) => {
    const records: any[] = []

    res.on("searchEntry", (entry: any) => {
      const formatted_entry = objectify_entry(entry)
      delete formatted_entry.userCertificate
      records.push(formatted_entry)
    })

    res.on("error", (err: any) => {
      reject(err)
    })

    res.on("end", () => {
      resolve(records)
    })
  })

export let client: ldap.Client

export const init = async () => {
  client = await connect()
  await bind(client)
}

export const query = async (opts: any) => {
  const searchresult = await search(client, opts)
  return await handle_search_response(searchresult)
}
