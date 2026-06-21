
import axios from 'axios'
import APIResponse from './base.apiresponse';

import { localStore as storage } from '@/utils/storage';

import log from '@/utils/log';
const logger = log.withPrefix("[api][base.ax]");

const defaultConfig = {
    // baseURL: "/",
    timeout: 200000,
    headers: {
      Accept: 'application/json',
    }
};

const LoginPath = '/';

/// !Axios Creation Functions
///========================================================
///========================================================
const axiosCreate = function(config) {
    const instance = axios.create({
        ...defaultConfig,
        ...config
    })

    instance.interceptors.response.use(
        function (response) {
            // Any status code within the 2xx range triggers this.
            // Wrap the raw Axios response into your APIResponse object.
            return new APIResponse(response);
        },
        function (error) {
            // Any status code outside the 2xx range triggers this.
            // If the server responded with an error (e.g., 400, 401, 500), 
            // error.response will exist.
            if (error.response) {

              logger.log("got error response", error.response)

                // Wrap the error response as well so your catch blocks get the same object structure
                return Promise.reject(new APIResponse(error.response));
            }

            // For network errors (like timeout or CORS) where no response was received
            return Promise.reject(error);
        }
    );

    return instance
}

const defaultAxios = function() {
  return axiosCreate({});
}

const loggedInAxios = function(token) {
  return axiosCreate({
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'application/json',
    }
  });
}


/// !AX Management
///========================================================
///========================================================
let AX;

const UpdateAX = {
  create(config) {
    AX = axiosCreate(config)
  },
  default() {
    AX = defaultAxios()
  },
  login(token) {
    storage.token = token;
    AX = loggedInAxios(token)
  },
  logout() {
    storage.token = undefined;
    AX = defaultAxios()
  }
}

if(storage.token) {
   AX = loggedInAxios(storage.token);
} else {
  UpdateAX.default();
}

/// !TokenExpiration
///========================================================
///========================================================
// status = FORBIDDEN
const UserNeedLogin = function(apiresponse) {
  return apiresponse.tokenExpired;
}

// let RedirectToLoginPageShown = false;
// const RedirectToLoginPage = function() {
//   if(!RedirectToLoginPageShown) {
//     RedirectToLoginPageShown = true;
//     notyUtil.noFurtherAction("Login session expired, you need to login again.", ()=>{
//       window.location.href = LoginPath;
//     }).show();
//   }
// }

// const RedirectToLoginPage = RunOnce(function() {
//     notyUtil.noFurtherAction("Login session expired, you need to login again.", ()=>{
//       window.location.href = LoginPath;
//     }).show();
// })

export {
  UserNeedLogin,
  UpdateAX,
  AX
}