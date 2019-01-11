import * as _ from 'lodash';
import * as types from './mutation-types';
import constants from './constants';

import discovery from '../../../api/discovery';
import api from '../../../api';

export default {
  setDiscoveryModel({ commit, state }, editorString) {
    const value = JSON.stringify(state.discoveryModel);
    const other = editorString;
    if (_.isEqual(value, other)) {
      return;
    }

    try {
      const discoveryModel = JSON.parse(editorString);
      commit(types.SET_DISCOVERY_MODEL, discoveryModel);
      commit(types.DISCOVERY_MODEL_PROBLEMS, null);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_TWO);
    } catch (e) {
      const problems = [{
        key: null,
        error: e.message,
      }];
      commit(types.DISCOVERY_MODEL_PROBLEMS, problems);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_ONE);
    }
  },
  setDiscoveryModelProblems({ commit }, problems) {
    commit(types.DISCOVERY_MODEL_PROBLEMS, problems);
    commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_TWO);
  },
  /**
   * Step 2: validate the Discovery Config.
   * Route: `/wizard/discovery-config`.
   */
  async validateDiscoveryConfig({ commit, state }) {
    try {
      const { success, problems } = await discovery.validateDiscoveryConfig(state.discoveryModel);
      if (success) {
        commit(types.DISCOVERY_MODEL_PROBLEMS, null);
        commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_THREE);
      } else {
        commit(types.DISCOVERY_MODEL_PROBLEMS, problems);
        commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_TWO);
      }
    } catch (e) {
      commit(types.DISCOVERY_MODEL_PROBLEMS, [{
        key: null,
        error: e.message,
      }]);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_TWO);
    }
    return null;
  },

  setConfigurationSigningPrivate({ commit, state }, signingPrivate) {
    if (_.isEqual(state.configuration.signing_private, signingPrivate)) {
      return;
    }

    commit(types.SET_CONFIGURATION_SIGNING_PRIVATE, signingPrivate);
    commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_THREE);
  },
  setConfigurationSigningPublic({ commit, state }, signingPublic) {
    if (_.isEqual(state.configuration.signing_public, signingPublic)) {
      return;
    }

    commit(types.SET_CONFIGURATION_SIGNING_PUBLIC, signingPublic);
    commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_THREE);
  },
  setConfigurationTransportPrivate({ commit, state }, transportPrivate) {
    if (_.isEqual(state.configuration.transport_private, transportPrivate)) {
      return;
    }

    commit(types.SET_CONFIGURATION_TRANSPORT_PRIVATE, transportPrivate);
    commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_THREE);
  },
  setConfigurationTransportPublic({ commit, state }, transportPublic) {
    if (_.isEqual(state.configuration.transport_public, transportPublic)) {
      return;
    }

    commit(types.SET_CONFIGURATION_TRANSPORT_PUBLIC, transportPublic);
    commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_THREE);
  },
  /**
   * Step 3: Validate the configuration.
   * Route: `/wizard/configuration`.
   */
  async validateConfiguration({ commit, state }) {
    commit(types.CLEAR_CONFIGURATION_ERRORS);

    if (_.isEmpty(state.configuration.signing_private)) {
      commit(types.ADD_CONFIGURATION_ERRORS, 'Signing Private Certificate (.key) empty');
    }
    if (_.isEmpty(state.configuration.signing_public)) {
      commit(types.ADD_CONFIGURATION_ERRORS, 'Signing Public Certificate (.pem) empty');
    }
    if (_.isEmpty(state.configuration.transport_private)) {
      commit(types.ADD_CONFIGURATION_ERRORS, 'Transport Private Certificate (.key) empty');
    }
    if (_.isEmpty(state.configuration.transport_public)) {
      commit(types.ADD_CONFIGURATION_ERRORS, 'Transport Public Certificate (.pem) empty');
    }

    if (!_.isEmpty(state.errors.configuration)) {
      return false;
    }

    try {
      // NB: We do not care what value this method call returns as long
      // as it does not throw, we know the configuration is valid.
      const { configuration } = state;
      await api.validateConfiguration(configuration);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_FOUR);

      return true;
    } catch (err) {
      commit(types.SET_CONFIGURATION_ERRORS, [err]);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_THREE);

      return false;
    }
  },
  /**
   * Sets array of Error objects.
   */
  setConfigurationErrors({ commit }, errors) {
    commit(types.SET_CONFIGURATION_ERRORS, errors);
  },
  /**
   * Step 4: Calls /api/test-cases to get all the test cases, then sets the
   * retrieved test cases in the store.
   * Route: `/wizard/run-overview`.
   */
  async computeTestCases({ commit }) {
    try {
      const testCases = await api.computeTestCases();
      commit(types.SET_TEST_CASES, testCases);
      commit(types.SET_TEST_CASES_ERROR, []);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_FIVE);
    } catch (err) {
      commit(types.SET_TEST_CASES, []);
      commit(types.SET_TEST_CASES_ERROR, [err]);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_FOUR);
    }
  },
  /**
   * Calls /api/report to get all the test cases, then sets the
   * retrieved test cases in the store.
   */
  async computeTestCaseResults({ commit }) {
    try {
      const testCaseResults = await api.computeTestCaseResults();
      commit(types.SET_TEST_CASE_RESULTS, testCaseResults);
      commit(types.SET_TEST_CASE_RESULTS_ERROR, []);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_SIX);
    } catch (err) {
      commit(types.SET_TEST_CASE_RESULTS, {});
      commit(types.SET_TEST_CASE_RESULTS_ERROR, [
        err,
      ]);
      commit(types.SET_WIZARD_STEP, constants.WIZARD.STEP_FIVE);
    }
  },
};
