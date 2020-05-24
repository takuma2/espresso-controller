import { createSlice, Dispatch, PayloadAction } from "@reduxjs/toolkit";
import moment from "moment";
import {
  GetTargetTemperatureRequest,
  GetTargetTemperatureResponse,
  SetTargetTemperatureRequest,
  SetTargetTemperatureResponse,
} from "../../proto/pkg/appliancepb/appliance_pb";
import { ServiceError } from "../../proto/pkg/appliancepb/appliance_pb_service";
import { applianceClient, createUnaryGrpcThunk } from "../helpers";

export const targetTemperatureSlice = createSlice({
  name: "targetTemp",
  initialState: {
    targetTemp: undefined,
    isFetching: false,
    isSetting: false,
  } as {
    targetTemp?: { value: number; setAt: moment.Moment };
    isFetching: boolean;
    isSetting: boolean;
  },
  reducers: {
    // GetTargetTemperature
    getTargetTemperatureRequest: (
      state,
      action: PayloadAction<GetTargetTemperatureRequest>
    ) => {
      state.isFetching = true;
    },
    getTargetTemperatureResponse: (
      state,
      action: PayloadAction<GetTargetTemperatureResponse>
    ) => {
      const value = action.payload.getTemperature();
      const setAt = action.payload.getSetAt()?.toDate();
      if (!setAt) {
        console.warn("Target temperature response missing setAt time");
        return;
      }

      state.targetTemp = { value, setAt: moment(setAt) };
      state.isFetching = false;
    },
    getTargetTemperatureFailure: (
      state,
      action: PayloadAction<{
        req: GetTargetTemperatureRequest;
        err: ServiceError;
      }>
    ) => {
      state.targetTemp = undefined;
      state.isFetching = false;
    },

    // SetTargetTemperature
    setTargetTemperatureRequest: (
      state,
      action: PayloadAction<SetTargetTemperatureRequest>
    ) => {
      state.isSetting = true;
    },
    setTargetTemperatureResponse: (
      state,
      action: PayloadAction<SetTargetTemperatureResponse>
    ) => {
      const value = action.payload.getTemperature();
      const setAt = action.payload.getSetAt()?.toDate();
      if (!setAt) {
        console.warn("Target temperature response missing setAt time");
        return;
      }

      state.targetTemp = { value, setAt: moment(setAt) };
      state.isSetting = false;
    },
    setTargetTemperatureFailure: (
      state,
      action: PayloadAction<{
        req: SetTargetTemperatureRequest;
        err: ServiceError;
      }>
    ) => {
      state.isSetting = false;
    },
  },
});

export const getTargetTemperature = (req: GetTargetTemperatureRequest) => (
  d: Dispatch
) =>
  createUnaryGrpcThunk(
    applianceClient.getTargetTemperature,
    req,
    {
      request: targetTemperatureSlice.actions.getTargetTemperatureRequest,
      response: targetTemperatureSlice.actions.getTargetTemperatureResponse,
      failure: targetTemperatureSlice.actions.getTargetTemperatureFailure,
    },
    d
  );

export const setTargetTemperature = (req: SetTargetTemperatureRequest) => (
  d: Dispatch
) =>
  createUnaryGrpcThunk(
    applianceClient.setTargetTemperature,
    req,
    {
      request: targetTemperatureSlice.actions.setTargetTemperatureRequest,
      response: targetTemperatureSlice.actions.setTargetTemperatureResponse,
      failure: targetTemperatureSlice.actions.setTargetTemperatureFailure,
    },
    d
  );
