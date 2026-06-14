
import axios from 'axios'

import { localStore as storage } from '@/utils/storage';

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
    return axios.create({
        ...defaultConfig,
        ...config
    })
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