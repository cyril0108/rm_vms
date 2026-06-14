// Storage

// Base storage constructor that works with
// both localStorage and sessionStorage

// import Logger from '@/logger'
// let logger = Logger.withPrefix("[Storage]");

const Storage = function(storage, stringProps, boolProps, jsonProps) {

    if ( !(this instanceof Storage) ) {

        throw new Error("Storage should be called with keyword 'new'.");
        // return new Storage(storage);

    }

    jsonProps.forEach(key=>{

        Object.defineProperty(this, key, {

            enumerable: true,

            get() {

                let json = storage.getItem(key);
                if(typeof json === "string") {
                    return JSON.parse(json);
                }
                return undefined;

            },

            set(val) {

                if(val === undefined) {
                    storage.removeItem(key);
                } else {
                    storage.setItem(key, JSON.stringify(val));
                }

            }

        });

    });

    stringProps.forEach(key=>{

        Object.defineProperty(this, key, {

            enumerable: true,

            get() {

                return storage.getItem(key);

            },

            set(val) {

                if(val === undefined) {
                    storage.removeItem(key);
                } else {
                    storage.setItem(key, val);
                }

            }

        });

    });

    boolProps.forEach(key=>{

        Object.defineProperty(this, key, {

            get() {

                return (storage.getItem(key) == "true");

            },

            set(val) {

                storage.setItem(key, !!val);

            }

        });

    });

    return this;

}

let localStore = new Storage(window.localStorage, ["token"], [], ["user"]);
let sessionStore = new Storage(window.sessionStorage, ["token"],[],[]);

export default Storage;

export {
    localStore,
    sessionStore
}