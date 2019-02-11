/**
 * This creates a real store so avoid having to mock things.
 * This makes testing much easier.
 *
 * See the recommendation:
 * https://vue-test-utils.vuejs.org/guides/using-with-vuex.html#testing-a-running-store
 */
import { createLocalVue } from '@vue/test-utils';
import Vuex from 'vuex';
import { cloneDeep } from 'lodash';

import actions from './actions';
import mutations from './mutations';
import getters from './getters';
import state from './state';
import * as types from './mutation-types';

import api from '../../../api';
// https://jestjs.io/docs/en/mock-functions#mocking-modules
jest.mock('../../../api');

describe('web/src/store/modules/config', () => {
  describe('config/configuration', () => {
    let dispatch;
    /**
     * Creates a real store so we don't have to mock things out.
     */
    const createRealStore = () => {
      const localVue = createLocalVue();
      localVue.use(Vuex);
      const store = new Vuex.Store({
        state: cloneDeep(state),
        actions,
        mutations,
        getters,
      });
      dispatch = jest.fn();
      store.dispatch = dispatch;

      return store;
    };

    it('configuration.{signing_private,signing_public,transport_private,transport_public, client_id, client_secret, token_endpoint, x_fapi_financial_id} initially empty', async () => {
      const store = createRealStore();

      expect(store.getters.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: '',
        transport_public: '',
        client_id: '',
        client_secret: '',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });
    });

    it('setConfigurationSigningPrivate', async () => {
      const store = createRealStore();

      const signingPrivate = 'signingPrivate';
      await actions.setConfigurationSigningPrivate(store, signingPrivate);

      expect(store.getters.configuration).toEqual({
        signing_private: signingPrivate,
        signing_public: '',
        transport_private: '',
        transport_public: '',
        client_id: '',
        client_secret: '',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });
    });

    it('setConfigurationSigningPublic', async () => {
      const store = createRealStore();

      const signingPublic = 'signingPublic';
      await actions.setConfigurationSigningPublic(store, signingPublic);

      expect(store.getters.configuration).toEqual({
        signing_private: '',
        signing_public: signingPublic,
        transport_private: '',
        transport_public: '',
        client_id: '',
        client_secret: '',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });
    });

    it('setConfigurationTransportPrivate', async () => {
      const store = createRealStore();

      const transportPrivate = 'transportPrivate';
      await actions.setConfigurationTransportPrivate(store, transportPrivate);

      expect(store.getters.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: transportPrivate,
        transport_public: '',
        client_id: '',
        client_secret: '',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });
    });

    it('setConfigurationTransportPublic', async () => {
      const store = createRealStore();

      const transportPublic = 'transportPublic';
      await actions.setConfigurationTransportPublic(store, transportPublic);

      expect(store.getters.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: '',
        transport_public: transportPublic,
        client_id: '',
        client_secret: '',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });
    });

    it('commmits client_id, client_secret, token_endpoint, x_fapi_financial_id and redirect_url', async () => {
      const store = createRealStore();

      expect(store.state.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: '',
        transport_public: '',
        client_id: '',
        client_secret: '',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });

      store.commit(types.SET_CLIENT_ID, '8672384e-9a33-439f-8924-67bb14340d71');
      expect(store.state.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: '',
        transport_public: '',
        client_id: '8672384e-9a33-439f-8924-67bb14340d71',
        client_secret: '',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });

      store.commit(types.SET_CLIENT_SECRET, '2cfb31a3-5443-4e65-b2bc-ef8e00266a77');
      expect(store.state.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: '',
        transport_public: '',
        client_id: '8672384e-9a33-439f-8924-67bb14340d71',
        client_secret: '2cfb31a3-5443-4e65-b2bc-ef8e00266a77',
        token_endpoint: '',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });

      store.commit(types.SET_TOKEN_ENDPOINT, 'https://modelobank2018.o3bank.co.uk:4201/token');
      expect(store.state.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: '',
        transport_public: '',
        client_id: '8672384e-9a33-439f-8924-67bb14340d71',
        client_secret: '2cfb31a3-5443-4e65-b2bc-ef8e00266a77',
        token_endpoint: 'https://modelobank2018.o3bank.co.uk:4201/token',
        x_fapi_financial_id: '',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });

      store.commit(types.SET_X_FAPI_FINANCIAL_ID, '0015800001041RHAAY');
      expect(store.state.configuration).toEqual({
        signing_private: '',
        signing_public: '',
        transport_private: '',
        transport_public: '',
        client_id: '8672384e-9a33-439f-8924-67bb14340d71',
        client_secret: '2cfb31a3-5443-4e65-b2bc-ef8e00266a77',
        token_endpoint: 'https://modelobank2018.o3bank.co.uk:4201/token',
        x_fapi_financial_id: '0015800001041RHAAY',
        redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
      });
    });

    describe('validateDiscoveryConfig', () => {
      afterEach(() => {
        jest.resetAllMocks();
      });

      it('commits token_endpoint after success', async () => {
        const store = createRealStore();

        expect(store.state.configuration).toEqual({
          signing_private: '',
          signing_public: '',
          transport_private: '',
          transport_public: '',
          client_id: '',
          client_secret: '',
          token_endpoint: '',
          x_fapi_financial_id: '',
          redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
        });

        api.validateDiscoveryConfig.mockReturnValueOnce({
          success: true,
          problems: [],
          response: {
            token_endpoints: {
              'schema_version=https://raw.githubusercontent.com/OpenBankingUK/read-write-api-specs/v3.0.0/dist/account-info-swagger.json': 'https://modelobank2018.o3bank.co.uk:4201/token_1',
              'schema_version=https://raw.githubusercontent.com/OpenBankingUK/read-write-api-specs/v3.0.0/dist/payment-initiation-swagger.json': 'https://modelobank2018.o3bank.co.uk:4201/token_2',
            },
          },
        });

        await actions.validateDiscoveryConfig(store);

        expect(store.state.configuration).toEqual({
          signing_private: '',
          signing_public: '',
          transport_private: '',
          transport_public: '',
          client_id: '',
          client_secret: '',
          token_endpoint: 'https://modelobank2018.o3bank.co.uk:4201/token_1',
          x_fapi_financial_id: '',
          redirect_url: 'https://0.0.0.0:8443/conformancesuite/callback',
        });
      });
    });

    describe('validateConfiguration', () => {
      afterEach(() => {
        jest.resetAllMocks();
      });

      it('setConfigurationSigningPrivate not called before validateConfiguration', async () => {
        const store = createRealStore();

        await actions.setConfigurationSigningPublic(store, 'setConfigurationSigningPublic');
        await actions.setConfigurationTransportPrivate(store, 'setConfigurationTransportPrivate');
        await actions.setConfigurationTransportPublic(store, 'setConfigurationTransportPublic');

        const valid = await actions.validateConfiguration(store);
        expect(valid).toEqual(false);

        const errors = [
          'Signing Private Certificate (.key) empty',
          'Client ID empty',
          'Client Secret empty',
          'Token Endpoint empty',
          'x-fapi-financial-id empty',
        ];
        expect(dispatch).toHaveBeenCalledWith('status/setErrors', errors, { root: true });
      });

      it('setConfigurationSigningPublic not called before validateConfiguration', async () => {
        const store = createRealStore();

        await actions.setConfigurationSigningPrivate(store, 'setConfigurationSigningPrivate');
        await actions.setConfigurationTransportPrivate(store, 'setConfigurationTransportPrivate');
        await actions.setConfigurationTransportPublic(store, 'setConfigurationTransportPublic');

        const valid = await actions.validateConfiguration(store);
        expect(valid).toEqual(false);

        const errors = [
          'Signing Public Certificate (.pem) empty',
          'Client ID empty',
          'Client Secret empty',
          'Token Endpoint empty',
          'x-fapi-financial-id empty',
        ];
        expect(dispatch).toHaveBeenCalledWith('status/setErrors', errors, { root: true });
      });

      it('setConfigurationTransportPrivate not called before validateConfiguration', async () => {
        const store = createRealStore();

        await actions.setConfigurationSigningPublic(store, 'setConfigurationSigningPublic');
        await actions.setConfigurationSigningPrivate(store, 'setConfigurationSigningPrivate');
        await actions.setConfigurationTransportPublic(store, 'setConfigurationTransportPublic');

        const valid = await actions.validateConfiguration(store);
        expect(valid).toEqual(false);

        const errors = [
          'Transport Private Certificate (.key) empty',
          'Client ID empty',
          'Client Secret empty',
          'Token Endpoint empty',
          'x-fapi-financial-id empty',
        ];
        expect(dispatch).toHaveBeenCalledWith('status/setErrors', errors, { root: true });
      });

      it('setConfigurationTransportPublic not called before validateConfiguration', async () => {
        const store = createRealStore();

        await actions.setConfigurationSigningPublic(store, 'setConfigurationSigningPublic');
        await actions.setConfigurationSigningPrivate(store, 'setConfigurationSigningPrivate');
        await actions.setConfigurationTransportPrivate(store, 'setConfigurationTransportPrivate');

        const valid = await actions.validateConfiguration(store);
        expect(valid).toEqual(false);

        const errors = [
          'Transport Public Certificate (.pem) empty',
          'Client ID empty',
          'Client Secret empty',
          'Token Endpoint empty',
          'x-fapi-financial-id empty',
        ];
        expect(dispatch).toHaveBeenCalledWith('status/setErrors', errors, { root: true });
      });

      it('setConfigurationSigningPrivate, setConfigurationSigningPublic, setConfigurationTransportPrivate and setConfigurationTransportPublic not called before validateConfiguration', async () => {
        const store = createRealStore();

        const valid = await actions.validateConfiguration(store);
        expect(valid).toEqual(false);

        const errors = [
          'Signing Private Certificate (.key) empty',
          'Signing Public Certificate (.pem) empty',
          'Transport Private Certificate (.key) empty',
          'Transport Public Certificate (.pem) empty',
          'Client ID empty',
          'Client Secret empty',
          'Token Endpoint empty',
          'x-fapi-financial-id empty',
        ];
        expect(dispatch).toHaveBeenCalledWith('status/setErrors', errors, { root: true });
      });

      it('setConfigurationSigningPrivate, setConfigurationSigningPublic, setConfigurationTransportPrivate and setConfigurationTransportPublic called before validateConfiguration', async () => {
        api.validateConfiguration.mockReturnValueOnce({
          signing_private: 'does_not_matter_what_the_value_is',
          signing_public: 'does_not_matter_what_the_value_is',
          transport_private: 'does_not_matter_what_the_value_is',
          transport_public: 'does_not_matter_what_the_value_is',
        });

        const store = createRealStore();
        store.commit(types.SET_CLIENT_ID, '8672384e-9a33-439f-8924-67bb14340d71');
        store.commit(types.SET_CLIENT_SECRET, '2cfb31a3-5443-4e65-b2bc-ef8e00266a77');
        store.commit(types.SET_TOKEN_ENDPOINT, 'https://modelobank2018.o3bank.co.uk:4201/token');
        store.commit(types.SET_X_FAPI_FINANCIAL_ID, '0015800001041RHAAY');

        await actions.setConfigurationSigningPublic(store, 'setConfigurationSigningPublic');
        await actions.setConfigurationSigningPrivate(store, 'setConfigurationSigningPrivate');
        await actions.setConfigurationTransportPrivate(store, 'setConfigurationTransportPrivate');
        await actions.setConfigurationTransportPublic(store, 'setConfigurationTransportPublic');

        const valid = await actions.validateConfiguration(store);
        expect(valid).toEqual(true);
      });

      it('setConfigurationSigningPrivate, setConfigurationSigningPublic, setConfigurationTransportPrivate and setConfigurationTransportPublic called with invalid values before validateConfiguration', async () => {
        const errorResponse = {
          error: "error with signing certificate: error with public key: asn1: structure error: tags don't match (16 vs {class:0 tag:2 length:1 isCompound:false}) {optional:false explicit:false application:false private:false defaultValue:\u003cnil\u003e tag:\u003cnil\u003e stringType:0 timeType:0 set:false omitEmpty:false} tbsCertificate @2",
        };
        api.validateConfiguration.mockRejectedValueOnce(errorResponse);

        const store = createRealStore();
        store.commit(types.SET_CLIENT_ID, '8672384e-9a33-439f-8924-67bb14340d71');
        store.commit(types.SET_CLIENT_SECRET, '2cfb31a3-5443-4e65-b2bc-ef8e00266a77');
        store.commit(types.SET_TOKEN_ENDPOINT, 'https://modelobank2018.o3bank.co.uk:4201/token');
        store.commit(types.SET_X_FAPI_FINANCIAL_ID, '0015800001041RHAAY');

        await actions.setConfigurationSigningPublic(store, 'not_a_certificate');
        await actions.setConfigurationSigningPrivate(store, 'not_a_certificate');
        await actions.setConfigurationTransportPrivate(store, 'not_a_certificate');
        await actions.setConfigurationTransportPublic(store, 'not_a_certificate');

        const valid = await actions.validateConfiguration(store);
        expect(valid).toEqual(false);

        expect(dispatch).toHaveBeenCalledWith('status/setErrors', [errorResponse], { root: true });
      });

      it('validateConfiguration clears previous errors', async () => {
        const store = createRealStore();

        // This will generate an error because we have not called any of the methods
        // that sets the values for the configuration.

        expect(await actions.validateConfiguration(store)).toEqual(false);
        const errors = [
          'Signing Private Certificate (.key) empty',
          'Signing Public Certificate (.pem) empty',
          'Transport Private Certificate (.key) empty',
          'Transport Public Certificate (.pem) empty',
          'Client ID empty',
          'Client Secret empty',
          'Token Endpoint empty',
          'x-fapi-financial-id empty',
        ];
        expect(dispatch).toHaveBeenCalledWith('status/setErrors', errors, { root: true });

        api.validateConfiguration.mockReturnValueOnce({
          signing_private: 'does_not_matter_what_the_value_is',
          signing_public: 'does_not_matter_what_the_value_is',
          transport_private: 'does_not_matter_what_the_value_is',
          transport_public: 'does_not_matter_what_the_value_is',
        });

        await actions.setConfigurationSigningPublic(store, 'setConfigurationSigningPublic');
        await actions.setConfigurationSigningPrivate(store, 'setConfigurationSigningPrivate');
        await actions.setConfigurationTransportPrivate(store, 'setConfigurationTransportPrivate');
        await actions.setConfigurationTransportPublic(store, 'setConfigurationTransportPublic');

        store.commit(types.SET_CLIENT_ID, '8672384e-9a33-439f-8924-67bb14340d71');
        store.commit(types.SET_CLIENT_SECRET, '2cfb31a3-5443-4e65-b2bc-ef8e00266a77');
        store.commit(types.SET_TOKEN_ENDPOINT, 'https://modelobank2018.o3bank.co.uk:4201/token');
        store.commit(types.SET_X_FAPI_FINANCIAL_ID, '0015800001041RHAAY');

        // This will clear out the previous errors, and will result in configurationErrors
        // being empty since they are not any errors.
        expect(await actions.validateConfiguration(store)).toEqual(true);
        expect(dispatch).toHaveBeenCalledWith('status/clearErrors', null, { root: true });
      });
    });
  });
});
